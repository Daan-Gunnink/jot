import type { JSONContent } from '@tiptap/vue-3'
import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { StateTree } from 'pinia'
import type { PersistenceOptions } from 'pinia-plugin-persistedstate'
import { v4 as uuidv4 } from 'uuid'

interface Jot {
  id: string
  revisionId: string
  title: string
  content: JSONContent
  createdAt: Date
  updatedAt: Date
}

// Type for our store state
interface JotState {
  jots: Jot[]
  revisionsMap: Map<string, Jot[]>
  redoMap: Map<string, Jot[]>
}

interface SerializedJot {
  id: string
  revisionId: string
  title: string
  content: string // Content is serialized as a string
  createdAt: string
  updatedAt: string
}

interface JotRevisionItem {
  id: string
  revisions: SerializedJot[]
}

export const useJotStore = defineStore(
  'jot',
  () => {
    const jots = ref<Jot[]>([])

    const createJot = (
      title: string = 'Untitled Jot',
      content?: JSONContent
    ): string => {
      const jot: Jot = {
        id: uuidv4(),
        revisionId: uuidv4(),
        title: title,
        content: content ?? { type: 'doc', content: [] },
        createdAt: new Date(),
        updatedAt: new Date(),
      }
      jots.value.push(jot)

      return jot.id
    }

    const updateJot = (
      id: string,
      title?: string,
      content?: JSONContent
    ): void => {
      const jot = jots.value.find((jot) => jot.id === id)
      if (jot) {
        jot.title = title ? limitTitleLength(title) : jot.title
        jot.content = content ?? jot.content
        jot.updatedAt = new Date()
      }
    }

    const deleteJot = (id: string): void => {
      jots.value = jots.value.filter((jot) => jot.id !== id)
    }

    const getJotById = (id: string): Jot | undefined => {
      return jots.value.find((jot) => jot.id === id)
    }

    const listJots = (): Jot[] => {
      return jots.value.sort(
        (a, b) => b.updatedAt.getTime() - a.updatedAt.getTime()
      )
    }

    return {
      jots,
      createJot,
      updateJot,
      deleteJot,
      getJotById,
      listJots,
    }
  },
  {
    // Using optimized persistence - only storing latest revision of each jot
    // This reduces storage requirements but limits undo/redo to the current session only
    persist: {
      serializer: {
        serialize: (state: StateTree): string => {
          const serializedJots = state.jots.map((jot: Jot) => {
            return {
                ...jot,
                content: JSON.stringify(jot.content),
                createdAt: jot.createdAt.toISOString(),
                updatedAt: jot.updatedAt.toISOString(),
            } as SerializedJot
          })

          return JSON.stringify({
            jots: serializedJots,
          })
        },
        deserialize: (serializedState: string): StateTree => {
          const parsedData = JSON.parse(serializedState)
          const jots : Jot[] = []

          // If latest revisions exist in the parsed data
          if (parsedData.jots) {
            // Loop through each entry in the latest revisions array
            parsedData.jots.forEach(
              (item: SerializedJot) => {
                // Convert the stringified content back to objects and dates back to Date objects
                const processedRevision = {
                  ...item,
                  content: JSON.parse(item.content), // Parse the stringified content
                  createdAt: new Date(item.createdAt),
                  updatedAt: new Date(item.updatedAt),
                }

                // Set the processed revision as the only revision for this id in the map
                jots.push(processedRevision)
              }
            )
          }

          return {
            ...parsedData,
            jots,
          }
        },
      },
    } as PersistenceOptions<JotState>,
  }
)

// Helper function to limit title length
function limitTitleLength(title: string, maxLength: number = 128): string {
  // If title is already shorter than max length, return it as is
  if ([...title].length <= maxLength) {
    return title
  }

  // Try to find the last whitespace within maxLength
  const titleSlice = [...title].slice(0, maxLength).join('')
  const lastWhitespaceIndex = titleSlice.lastIndexOf(' ')

  // If whitespace found, break at that point
  if (lastWhitespaceIndex > 0) {
    return titleSlice.substring(0, lastWhitespaceIndex) + '...'
  }

  // If no whitespace found, simply truncate and add ellipsis
  return titleSlice + '...'
}
