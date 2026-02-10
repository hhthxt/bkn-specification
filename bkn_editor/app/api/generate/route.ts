import { NextRequest } from 'next/server';
import OpenAI from 'openai';
import Anthropic from '@anthropic-ai/sdk';

// Build system prompt with BKN specification summary
function buildSystemPrompt(
  dataSourcesSummary: string,
  existingFiles: Record<string, string>,
  currentFile?: string
): string {
  const existingFilesContext = Object.entries(existingFiles)
    .map(([path, content]) => `\n\n## File: ${path}\n\`\`\`markdown\n${content}\n\`\`\``)
    .join('\n');

  const currentFileContext = currentFile
    ? `\n\n## Current File Being Edited\n\`\`\`markdown\n${currentFile}\n\`\`\``
    : '';

  return `You are a BKN (Business Knowledge Network) expert. Your task is to generate valid BKN Markdown content based on user requests.

## BKN Format Rules

### File Structure
Each BKN file has two parts:
1. YAML Frontmatter (metadata)
2. Markdown Body (content)

### Frontmatter Types
- \`type: entity\` - Single entity definition
- \`type: relation\` - Single relation definition  
- \`type: action\` - Single action definition
- \`type: network\` - Complete network with multiple definitions
- \`type: fragment\` - Mixed fragment with multiple types

### Entity Format
\`\`\`markdown
---
type: entity
id: {entity_id}
name: {实体名称}
network: {network_id}
---

## Entity: {entity_id}

**{显示名称}** - {简短描述}

### Data Source

| Type | ID | Name |
|------|-----|------|
| data_view | {view_id} | {view_name} |

> **Primary Key**: \`{primary_key}\` | **Display Key**: \`{display_key}\`

### Property Override

| Property | Display Name | Index Config | Description |
|----------|--------------|--------------|-------------|
| {property} | {显示名} | {索引配置} | {说明} |

### Data Properties

| Property | Display Name | Type | Description | Primary Key | Index |
|----------|--------------|------|--------------|:-----------:|:-----:|
| {name} | {display} | {type} | {desc} | YES/NO | YES/NO |

### Logic Properties

#### {property_name}

- **类型**: metric | operator
- **来源**: {source_id} ({source_type})
- **说明**: {description}

| Parameter | Source | Binding |
|-----------|--------|---------|
| {param} | property | {property_name} |
| {param} | input | - |
\`\`\`

### Relation Format
\`\`\`markdown
---
type: relation
id: {relation_id}
name: {关系名称}
network: {network_id}
---

## Relation: {relation_id}

**{显示名称}** - {简短描述}

| Source | Target | Type |
|--------|--------|------|
| {source_entity} | {target_entity} | direct |

### Mapping Rules

| Source Property | Target Property |
|-----------------|-----------------|
| {source_prop} | {target_prop} |
\`\`\`

### Action Format
\`\`\`markdown
---
type: action
id: {action_id}
name: {行动名称}
network: {network_id}
action_type: add | modify | delete
---

## Action: {action_id}

**{显示名称}** - {简短描述}

| Bound Entity | Action Type |
|--------------|-------------|
| {entity_id} | modify |

### Trigger Condition

\`\`\`yaml
condition:
  object_type_id: {entity_id}
  field: {property_name}
  operation: == | != | > | < | >= | <= | in | not_in
  value: {value}
\`\`\`

### Tool Configuration

| Type | Toolbox ID | Tool ID |
|------|------------|---------|
| tool | {box_id} | {tool_id} |

### Parameter Binding

| Parameter | Source | Binding | Description |
|-----------|--------|---------|-------------|
| {param} | property | {property_name} | {说明} |
| {param} | input | - | {说明} |
| {param} | const | {value} | {说明} |
\`\`\`

## Available Data Sources

${dataSourcesSummary}

## Existing Project Files

${existingFilesContext}${currentFileContext}

## Important Rules

1. Output ONLY the BKN Markdown content (including YAML frontmatter)
2. Do NOT include code fences (\`\`\`markdown) around the output
3. Use valid entity/relation IDs that exist in the project when referencing them
4. Follow the exact table formats shown above
5. Use Chinese for display names and descriptions unless specified otherwise
6. Ensure all required fields are present
7. Use consistent naming conventions (lowercase with underscores for IDs)`;
}

type AIProvider = 'openai' | 'anthropic';

function getProvider(): AIProvider {
  const provider = (process.env.AI_PROVIDER || 'openai').toLowerCase();
  return provider === 'anthropic' ? 'anthropic' : 'openai';
}

export async function POST(request: NextRequest) {
  try {
    const provider = getProvider();

    if (provider === 'openai') {
      if (!process.env.OPENAI_API_KEY) {
        return new Response(
          JSON.stringify({ error: 'OPENAI_API_KEY not configured. Set OPENAI_API_KEY in .env.local or use AI_PROVIDER=anthropic with ANTHROPIC_API_KEY.' }),
          { status: 500, headers: { 'Content-Type': 'application/json' } }
        );
      }
    } else {
      if (!process.env.ANTHROPIC_API_KEY) {
        return new Response(
          JSON.stringify({ error: 'ANTHROPIC_API_KEY not configured. Set ANTHROPIC_API_KEY in .env.local or use AI_PROVIDER=openai with OPENAI_API_KEY.' }),
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
      const openai = new OpenAI({ apiKey: process.env.OPENAI_API_KEY });
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
