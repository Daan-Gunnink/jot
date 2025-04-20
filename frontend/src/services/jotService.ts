import { db, type Jot } from '../db';
import { extractTextFromTipTap } from '../utils/tiptapUtils';
import type { JSONContent } from '@tiptap/vue-3';
import { liveQuery } from 'dexie';
import { useObservable } from "@vueuse/rxjs";
import { from } from 'rxjs';

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

  await db.transaction('rw', db.jots, async () => {
    await db.jots.add(newJot);
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

  await db.transaction('rw', db.jots, async () => {
    await db.jots.update(id, updatedJot);
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
  await db.transaction('rw', db.jots, async () => {
    await db.jots.delete(id);
  });
}


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