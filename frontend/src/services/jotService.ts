import { db, type Jot, type SearchIndexEntry } from '../db';
import { extractTextFromTipTap } from '../utils/tiptapUtils';
import type { JSONContent } from '@tiptap/vue-3';
import { liveQuery } from 'dexie';
import { useObservable } from "@vueuse/rxjs";
import { from } from 'rxjs';

// --- Tokenization ---

/**
 * Simple tokenizer: converts to lowercase, splits by non-alphanumeric,
 * filters out empty strings and short words.
 * TODO: Consider adding stop word filtering.
 * @param text The text to tokenize.
 * @returns An array of unique words (tokens).
 */
function tokenize(text: string): string[] {
  if (!text) return [];
  const words = text
    .toLowerCase()
    .split(/[^a-z0-9]+/)
    .filter(word => word.length > 2); // Ignore short words (e.g., <= 2 chars)
  return [...new Set(words)]; // Return unique words
}

// --- Search Index Management ---

/**
 * Updates the search index for a given Jot.
 * Clears existing entries and adds new ones based on title and textContent.
 * @param jotId The ID of the Jot.
 * @param title The title of the Jot.
 * @param textContent The plain text content of the Jot.
 */
async function updateSearchIndex(jotId: string, title: string, textContent: string): Promise<void> {
  // Use a transaction for atomicity
  await db.transaction('rw', db.searchIndex, async () => {
    // 1. Clear existing index entries for this jotId
    await db.searchIndex.where({ jotId }).delete();

    // 2. Tokenize title and content
    const titleTokens = tokenize(title);
    const contentTokens = tokenize(textContent);
    const allTokens = [...new Set([...titleTokens, ...contentTokens])]; // Combine and ensure uniqueness

    // 3. Prepare new index entries
    const entries: SearchIndexEntry[] = allTokens.map(word => ({
      word,
      jotId,
    }));

    // 4. Bulk add new entries
    if (entries.length > 0) {
      await db.searchIndex.bulkAdd(entries);
    }
  });
}

// --- CRUD Operations ---

/**
 * Adds a new Jot to the database and updates the search index.
 * @param jotData Object containing title and content.
 * @param id The UUID for the new Jot.
 * @returns The newly created Jot object.
 */
export async function addJot(jotData: { title: string; content: JSONContent }, id: string): Promise<Jot> {
  const textContent = extractTextFromTipTap(jotData.content);
  const now = new Date();

  const newJot: Jot = {
    ...jotData,
    id,
    textContent,
    createdAt: now,
    updatedAt: now,
  };

  await db.transaction('rw', db.jots, db.searchIndex, async () => {
    await db.jots.add(newJot);
    await updateSearchIndex(newJot.id, newJot.title, newJot.textContent);
  });

  return newJot;
}

/**
 * Updates an existing Jot in the database and its search index.
 * @param id The ID of the Jot to update.
 * @param updateData Object containing optional title and content updates.
 * @returns The updated Jot object or null if not found or ID is invalid.
 */
export async function updateJot(id: string | null | undefined, updateData: { title?: string; content?: JSONContent }): Promise<Jot | null> {
  // Add check for valid ID before querying Dexie
  if (typeof id !== 'string' || id === '') {
    console.warn('updateJot called with invalid ID:', id);
    return null;
  }

  // Now we know id is a valid string, proceed with the first get
  const jot = await db.jots.get(id);
  if (!jot) return null;

  let needsIndexUpdate = false;
  const updatedJot: Partial<Jot> = { updatedAt: new Date() };

  if (updateData.title !== undefined && updateData.title !== jot.title) {
    updatedJot.title = updateData.title;
    needsIndexUpdate = true;
  }

  let newTextContent = jot.textContent;
  if (updateData.content !== undefined) {
    const calculatedTextContent = extractTextFromTipTap(updateData.content);
    if (calculatedTextContent !== jot.textContent) {
      updatedJot.content = updateData.content;
      updatedJot.textContent = calculatedTextContent;
      newTextContent = calculatedTextContent;
      needsIndexUpdate = true;
    }
  }

  await db.transaction('rw', db.jots, db.searchIndex, async () => {
    await db.jots.update(id, updatedJot);
    if (needsIndexUpdate) {
      await updateSearchIndex(id, updatedJot.title ?? jot.title, newTextContent);
    }
  });

  // Return the full updated jot
  const updatedJotResult = await db.jots.get(id); // ID is known to be valid here
  return updatedJotResult ?? null;
}

/**
 * Deletes a Jot from the database and removes its entries from the search index.
 * @param id The ID of the Jot to delete.
 */
export async function deleteJot(id: string): Promise<void> {
  await db.transaction('rw', db.jots, db.searchIndex, async () => {
    await db.jots.delete(id);
    await db.searchIndex.where({ jotId: id }).delete();
  });
}

// --- Read Operations ---

/**
 * Retrieves a single Jot by its ID.
 * @param id The ID of the Jot.
 * @returns The Jot object or undefined if not found or ID is invalid.
 */
export async function getJotById(id: string | null | undefined): Promise<Jot | undefined> {
  // Add check for valid ID before querying Dexie
  if (typeof id !== 'string' || id === '') {
      console.warn('getJotById called with invalid ID:', id);
      return undefined;
  }
  return db.jots.get(id);
}

/**
 * Provides a reactive list of all Jots, sorted by updated date (descending).
 * Uses Dexie's liveQuery and @vueuse/rxjs for Vue reactivity.
 */
export function listJotsReactive() {
    return useObservable(
        from(
            liveQuery(() => db.jots.orderBy('updatedAt').reverse().toArray())
        )
    );
}

/**
 * Gets the most recently updated Jot.
 * @returns The latest Jot or undefined if the database is empty.
 */
export async function getLatestJot(): Promise<Jot | undefined> {
    return db.jots.orderBy('updatedAt').reverse().first();
}

// --- Search Function ---

/**
 * Searches for Jots based on a query string.
 * Tokenizes the query and finds matching Jots via the searchIndex table.
 * @param query The search query string.
 * @returns A promise resolving to an array of matching Jot objects, potentially ranked or sorted.
 */
export async function searchJots(query: string): Promise<Jot[]> {
  const queryTokens = tokenize(query);

  if (queryTokens.length === 0) {
    return []; // No valid tokens to search for
  }

  // Find all index entries matching any of the query tokens
  const matchingIndexEntries = await db.searchIndex
    .where('word')
    .anyOf(queryTokens)
    .toArray();

  if (matchingIndexEntries.length === 0) {
    return []; // No matches found
  }

  // Extract unique Jot IDs
  const jotIds = [...new Set(matchingIndexEntries.map(entry => entry.jotId))];

  // --- Basic Ranking (Optional but Recommended) ---
  // Count how many query tokens match for each jotId
  const jotScores: { [id: string]: number } = {};
  matchingIndexEntries.forEach(entry => {
    jotScores[entry.jotId] = (jotScores[entry.jotId] || 0) + 1;
  });

  // Sort jotIds by score (descending)
  const sortedJotIds = jotIds.sort((a, b) => (jotScores[b] || 0) - (jotScores[a] || 0));

  // Fetch the full Jot objects for the matching IDs
  // Use bulkGet for efficiency
  const jots = await db.jots.bulkGet(sortedJotIds);

  // Filter out any undefined results (if a Jot was deleted but index wasn't updated somehow)
  // and return in the order determined by ranking
  return jots.filter((jot): jot is Jot => jot !== undefined);

  // --- Alternative: No Ranking (Simpler) ---
  /*
  const jots = await db.jots.where('id').anyOf(jotIds).toArray();
  // Maybe sort by updatedAt as a default?
  return jots.sort((a, b) => b.updatedAt.getTime() - a.updatedAt.getTime());
  */
} 