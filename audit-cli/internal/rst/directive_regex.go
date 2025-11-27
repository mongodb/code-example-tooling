package rst

import "regexp"

// RST Directive Regular Expressions
//
// This file contains all regular expressions for matching RST directives.
// These patterns are shared across the codebase to ensure consistency.

// IncludeDirectiveRegex matches .. include:: directives in RST files.
// Example: .. include:: /path/to/file.rst
var IncludeDirectiveRegex = regexp.MustCompile(`^\.\.\s+include::\s+(.+)$`)

// LiteralIncludeDirectiveRegex matches .. literalinclude:: directives in RST files.
// Example: .. literalinclude:: /path/to/file.py
var LiteralIncludeDirectiveRegex = regexp.MustCompile(`^\.\.\s+literalinclude::\s+(.+)$`)

// IOCodeBlockDirectiveRegex matches .. io-code-block:: directives in RST files.
// Example: .. io-code-block::
var IOCodeBlockDirectiveRegex = regexp.MustCompile(`^\.\.\s+io-code-block::`)

// InputDirectiveRegex matches .. input:: directives within io-code-block in RST files.
// Example: .. input:: /path/to/file.js
var InputDirectiveRegex = regexp.MustCompile(`^\.\.\s+input::\s+(.+)$`)

// OutputDirectiveRegex matches .. output:: directives within io-code-block in RST files.
// Example: .. output:: /path/to/file.json
var OutputDirectiveRegex = regexp.MustCompile(`^\.\.\s+output::\s+(.+)$`)

// ToctreeDirectiveRegex matches .. toctree:: directives in RST files.
// Example: .. toctree::
var ToctreeDirectiveRegex = regexp.MustCompile(`^\.\.\s+toctree::`)

// ProcedureDirectiveRegex matches .. procedure:: directives in RST files.
// Example: .. procedure::
var ProcedureDirectiveRegex = regexp.MustCompile(`^\.\.\s+procedure::`)

// StepDirectiveRegex matches .. step:: directives in RST files.
// Example: .. step:: Connect to the database
var StepDirectiveRegex = regexp.MustCompile(`^\.\.\s+step::\s*(.*)$`)

// TabsDirectiveRegex matches .. tabs:: directives in RST files.
// Example: .. tabs::
var TabsDirectiveRegex = regexp.MustCompile(`^\.\.\s+tabs::`)

// TabDirectiveRegex matches .. tab:: directives in RST files.
// Example: .. tab:: Python
var TabDirectiveRegex = regexp.MustCompile(`^\.\.\s+tab::\s*(.*)$`)

// ComposableTutorialDirectiveRegex matches .. composable-tutorial:: directives in RST files.
// Example: .. composable-tutorial::
var ComposableTutorialDirectiveRegex = regexp.MustCompile(`^\.\.\s+composable-tutorial::`)

// SelectedContentDirectiveRegex matches .. selected-content:: directives in RST files.
// Example: .. selected-content::
var SelectedContentDirectiveRegex = regexp.MustCompile(`^\.\.\s+selected-content::`)

