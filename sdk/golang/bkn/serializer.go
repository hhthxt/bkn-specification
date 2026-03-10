// Copyright The kweaver-ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package bkn

// Serialize is deprecated. Use the specific serialize functions in tar_writer.go instead.
// This function is kept for backward compatibility but may be removed in future versions.
func Serialize(doc *BknNetwork) string {
	// For backward compatibility, serialize only the frontmatter
	return serializeFrontmatter(doc.BknNetworkFrontmatter)
}
