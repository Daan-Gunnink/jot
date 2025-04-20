import type { JSONContent } from "@tiptap/vue-3";

/**
 * Recursively extracts plain text from TipTap JSONContent.
 * Traverses the node tree and concatenates text content.
 *
 * @param node The JSONContent node to process.
 * @returns The extracted plain text string.
 */
export function extractTextFromTipTap(
  node: JSONContent | undefined | null,
): string {
  if (!node) {
    return "";
  }

  let text = "";

  if (node.type === "text" && node.text) {
    text += node.text;
  }

  if (node.content && Array.isArray(node.content)) {
    node.content.forEach((childNode, index) => {
      text += extractTextFromTipTap(childNode);
      // Add a space between content blocks for better word separation during tokenization
      // Avoid adding a space after the very last node
      if (index < node.content!.length - 1) {
        text += " ";
      }
    });
  }

  // Add a space after block elements like paragraphs or headings if they aren't the last element in their parent
  // This helps separate words that might otherwise run together.
  if (
    node.type &&
    ["paragraph", "heading", "listItem", "blockquote", "codeBlock"].includes(
      node.type,
    )
  ) {
    // This check needs context from the parent, difficult here.
    // The space added within the loop might be sufficient for most cases.
    // Consider adding a newline instead if structure is important, but for search, spaces are often enough.
    // text += ' '; // Simplified approach: add space after processing children
  }

  return text.replace(/\s+/g, " ").trim(); // Consolidate multiple spaces and trim ends
}
