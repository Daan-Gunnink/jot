import Dexie, { type Table } from 'dexie';
import type { JSONContent } from '@tiptap/vue-3';

// Define the structure of a Jot item as stored in Dexie
export interface Jot {
  id: string; // Primary key (UUID from jotStore)
  title: string;
  content: JSONContent; // Store the original TipTap JSON
  textContent: string; // Store the searchable plain text version
  createdAt: Date;
  updatedAt: Date;
}

// Define the structure for the search index entries
export interface SearchIndexEntry {
  id?: number; // Optional auto-incrementing primary key for this table
  word: string; // The indexed word (token)
  jotId: string; // Foreign key linking back to the Jot item's id
}

export class JotDatabase extends Dexie {
  // Declare tables
  jots!: Table<Jot, string>; // Primary key is string (the UUID)
  searchIndex!: Table<SearchIndexEntry, number>; // Primary key is number (auto-incrementing)

  constructor() {
    super('JotDatabase'); // Database name
    this.version(1).stores({
      // Schema definition for version 1
      jots: '&id, title, updatedAt, textContent', // &id = Primary key, unique. Index title, updatedAt. textContent is stored but not indexed here directly.
      searchIndex: '++id, word, jotId, [word+jotId]', // ++id = Auto-incrementing PK. Index word, jotId. Compound index for efficient lookup and uniqueness.
    });
  }
}

// Create a singleton instance of the database
export const db = new JotDatabase(); 