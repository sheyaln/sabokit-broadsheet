/**
 * Notifuse Editor - Public API
 *
 * Main entry point for the blog editor with dynamic styling support
 */

// Main component
export { BroadsheetEditor, DEFAULT_INITIAL_CONTENT } from './BroadsheetEditor'
export type { BroadsheetEditorProps, BroadsheetEditorRef, TOCAnchor } from './BroadsheetEditor'

// Types
export type {
  EditorStyleConfig,
  CSSValue,
  DefaultStyles,
  ParagraphStyles,
  HeadingStyles,
  HeadingLevelStyles,
  CaptionStyles,
  SeparatorStyles,
  CodeBlockStyles,
  BlockquoteStyles,
  InlineCodeStyles,
  ListStyles,
  LinkStyles
} from './types/EditorStyleConfig'

// Default configuration
export { defaultEditorStyles } from './config/defaultEditorStyles'

// Style presets
export {
  academicPaperPreset
} from './presets'

// Utility functions
export { generateBlogPostCSS, clearCSSCache } from './utils/styleUtils'
export { validateStyleConfig, StyleConfigValidationError } from './utils/validateStyleConfig'
