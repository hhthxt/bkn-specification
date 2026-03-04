import { NextRequest } from 'next/server';
import { readFileSync, existsSync } from 'fs';
import { join } from 'path';
import OpenAI from 'openai';
import Anthropic from '@anthropic-ai/sdk';

const LLM_SPEC_PATH = join(process.cwd(), '..', 'docs', 'SPECIFICATION.llm.md');

function loadSpecForLLM(): string {
  if (existsSync(LLM_SPEC_PATH)) {
    return readFileSync(LLM_SPEC_PATH, 'utf-8');
  }
  return `# BKN 规范
生成符合 BKN 格式的 Markdown。文件结构：YAML Frontmatter + Markdown Body。
类型：object / relation / action。ID 小写+下划线，显示名用中文。
输出仅含 BKN 内容，勿用 \`\`\`markdown 包裹。`;
}

// Build system prompt with BKN specification and context
function buildSystemPrompt(
  dataSourcesSummary: string,
  existingFiles: Record<string, string>,
  currentFile?: string
): string {
  const spec = loadSpecForLLM();

  const existingFilesContext = Object.entries(existingFiles)
    .map(([path, content]) => `\n\n## 项目文件: ${path}\n\`\`\`markdown\n${content}\n\`\`\``)
    .join('\n');

  const currentFileContext = currentFile
    ? `\n\n## 当前正在编辑的文件\n\`\`\`markdown\n${currentFile}\n\`\`\``
    : '';

  return `${spec}

---

## 当前上下文

### 可用数据源
${dataSourcesSummary || '（无）'}

### 已有项目文件
${existingFilesContext || '（无）'}${currentFileContext}

### 生成要求
根据用户请求和上述规范生成 BKN 内容。引用对象/关系时使用项目中已有的 ID。`;
}

type AIProvider = 'openai' | 'anthropic';

function getProvider(): AIProvider {
  const provider = (process.env.AI_PROVIDER || 'openai').toLowerCase();
  return provider === 'anthropic' ? 'anthropic' : 'openai';
}

function isPlaceholderOrEmpty(key: string | undefined): boolean {
  if (!key || !key.trim()) return true;
  const placeholder = /^sk-your-api-key-here$/i;
  return placeholder.test(key.trim());
}

export async function POST(request: NextRequest) {
  try {
    const provider = getProvider();

    if (provider === 'openai') {
      if (isPlaceholderOrEmpty(process.env.OPENAI_API_KEY)) {
        return new Response(
          JSON.stringify({
            error: 'OPENAI_API_KEY not configured. Set OPENAI_API_KEY in bkn_editor/.env.local with a valid key, or set AI_PROVIDER=anthropic to use Anthropic.',
          }),
          { status: 500, headers: { 'Content-Type': 'application/json' } }
        );
      }
    } else {
      if (isPlaceholderOrEmpty(process.env.ANTHROPIC_API_KEY)) {
        return new Response(
          JSON.stringify({
            error: 'ANTHROPIC_API_KEY not configured. Set ANTHROPIC_API_KEY in bkn_editor/.env.local with a valid key, or set AI_PROVIDER=openai to use OpenAI.',
          }),
          { status: 500, headers: { 'Content-Type': 'application/json' } }
        );
      }
    }

    const body = await request.json();
    const { prompt, context } = body;

    if (!prompt) {
      return new Response(
        JSON.stringify({ error: 'Prompt is required' }),
        { status: 400, headers: { 'Content-Type': 'application/json' } }
      );
    }

    const systemPrompt = buildSystemPrompt(
      context?.dataSourcesSummary || '',
      context?.existingFiles || {},
      context?.currentFile
    );

    const encoder = new TextEncoder();

    const requestSignal = request.signal;

    if (provider === 'openai') {
      const openai = new OpenAI({
        apiKey: process.env.OPENAI_API_KEY,
        ...(process.env.OPENAI_BASE_URL && { baseURL: process.env.OPENAI_BASE_URL }),
      });
      const stream = await openai.chat.completions.create({
        model: process.env.OPENAI_MODEL || 'gpt-4o-mini',
        messages: [
          { role: 'system', content: systemPrompt },
          { role: 'user', content: prompt },
        ],
        stream: true,
        temperature: 0.7,
        signal: requestSignal,
      });

      const readableStream = new ReadableStream({
        async start(controller) {
          try {
            for await (const chunk of stream) {
              const content = chunk.choices[0]?.delta?.content || '';
              if (content) {
                controller.enqueue(encoder.encode(content));
              }
            }
            controller.close();
          } catch (error) {
            controller.error(error);
          }
        },
      });

      return new Response(readableStream, {
        headers: {
          'Content-Type': 'text/plain; charset=utf-8',
          'Cache-Control': 'no-cache',
          'Connection': 'keep-alive',
        },
      });
    }

    // Anthropic
    const anthropic = new Anthropic({
      apiKey: process.env.ANTHROPIC_API_KEY,
      ...(process.env.ANTHROPIC_BASE_URL && { baseURL: process.env.ANTHROPIC_BASE_URL }),
    });
    const messageStream = anthropic.messages.stream({
      model: process.env.ANTHROPIC_MODEL || 'claude-sonnet-4-20250514',
      max_tokens: 4096,
      system: systemPrompt,
      messages: [{ role: 'user', content: prompt }],
    });

    // Abort Anthropic stream when client disconnects
    requestSignal?.addEventListener('abort', () => messageStream.abort());

    const readableStream = new ReadableStream({
      async start(controller) {
        try {
          messageStream.on('text', (textDelta: string) => {
            if (textDelta) {
              controller.enqueue(encoder.encode(textDelta));
            }
          });
          messageStream.on('error', (err) => {
            controller.error(err);
          });
          await messageStream.done();
          controller.close();
        } catch (error) {
          controller.error(error);
        }
      },
    });

    return new Response(readableStream, {
      headers: {
        'Content-Type': 'text/plain; charset=utf-8',
        'Cache-Control': 'no-cache',
        'Connection': 'keep-alive',
      },
    });
  } catch (error: unknown) {
    const err = error as Error;
    console.error('Error generating BKN content:', err);
    return new Response(
      JSON.stringify({ error: err.message || 'Failed to generate content' }),
      { status: 500, headers: { 'Content-Type': 'application/json' } }
    );
  }
}
