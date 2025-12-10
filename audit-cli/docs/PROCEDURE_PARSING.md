# Procedure Parsing - Business Logic and Design Decisions

This document describes the business logic behind procedure parsing in the `audit-cli` tool. It explains what constitutes a procedure, how variations are detected, and key design decisions that govern the parser's behavior.

## Table of Contents

- [Overview](#overview)
- [What is a Procedure?](#what-is-a-procedure)
- [Procedure Formats](#procedure-formats)
- [Procedure Variations](#procedure-variations)
- [Sub-Procedures and List Type Tracking](#sub-procedures-and-list-type-tracking)
- [Include Directive Handling](#include-directive-handling)
- [Uniqueness and Grouping](#uniqueness-and-grouping)
- [Analysis vs. Extraction Semantics](#analysis-vs-extraction-semantics)
- [Key Design Decisions](#key-design-decisions)
- [Common Patterns and Edge Cases](#common-patterns-and-edge-cases)

## Overview

The procedure parser (`internal/rst/procedure_parser.go`) extracts and analyzes procedural content from MongoDB's reStructuredText (RST) documentation. MongoDB documentation uses procedures inconsistently across different contexts (drivers, deployment methods, platforms, etc.), so the parser must handle multiple formats and variation mechanisms.

## What is a Procedure?

A **procedure** is a set of sequential steps that guide users through a task. Examples include:
- Installing MongoDB
- Connecting to a cluster
- Creating a database
- Deploying an application

Procedures have:
- A **title/heading** (the section heading above the procedure)
- A **series of steps** (numbered or bulleted instructions)
- Optional **variations** (different content for different contexts)
- Optional **sub-procedures** (ordered lists within steps, each tracked separately with its list marker type)

## Procedure Formats

MongoDB documentation uses three formats for procedures:

### 1. Procedure Directive

The most common format uses `.. procedure::` and `.. step::` directives:

```rst
Before You Begin
----------------

.. procedure::

   .. step:: Create a MongoDB Atlas account

      Navigate to the MongoDB Atlas website and sign up for a free account.

   .. step:: Create a cluster

      Click "Build a Cluster" and select the free tier.

   .. step:: Configure network access

      Add your IP address to the IP Access List.
```

### 2. Ordered Lists

Some procedures use simple numbered or lettered lists:

```rst
Installation Steps
------------------

1. Download the MongoDB installer from the official website.

2. Run the installer and follow the prompts.

3. Verify the installation by running ``mongod --version``.
```

Or with letters:

```rst
a. First step
b. Second step
c. Third step
```

**Continuation Markers:** MongoDB documentation uses `#.` as a continuation marker for ordered lists, allowing the build system to automatically number items:

```rst
a. First step

#. Second step (automatically becomes 'b.')

#. Third step (automatically becomes 'c.')
```

The parser recognizes `#.` as a continuation of the current list type (numbered or lettered) and converts it to the appropriate next marker.

### 2a. Hierarchical Procedures with Numbered Headings

Some procedures use numbered headings to represent top-level steps, with ordered lists as sub-steps:

```rst
Procedure
---------

1. Modify the Keyfile
~~~~~~~~~~~~~~~~~~~~~

Update the keyfile to include both old and new keys.

a. Open the keyfile in a text editor.

#. Add the new key on a separate line.

#. Save the file.

2. Restart Each Member
~~~~~~~~~~~~~~~~~~~~~~

Restart all members one at a time.

a. Shut down the member.

#. Restart the member.
```

**Parser Behavior:**
- Detects "Procedure" heading followed by numbered headings (1., 2., 3., etc.)
- Treats numbered headings as top-level steps of a single procedure
- Parses ordered lists within each numbered heading as sub-steps
- Sets `HasSubSteps` flag to true if sub-steps are found
- **Analysis:** Shows 1 procedure with N steps (where N is the number of numbered headings)
- **Extraction:** Creates 1 file containing all numbered heading steps and their sub-steps

### 3. YAML Steps Files

MongoDB's build system converts YAML files to procedures:

```yaml
title: Connect to MongoDB
steps:
  - step: Import the MongoDB client
    content: |
      Import the MongoClient class from the pymongo package.
  - step: Create a connection string
    content: |
      Define your connection string with your credentials.
```

The parser detects references to these YAML files and extracts the steps.

## Procedure Variations

MongoDB documentation represents the same logical procedure differently for different contexts (Node.js vs. Python, Atlas CLI vs. drivers, macOS vs. Windows, etc.). The parser handles three mechanisms for variations:

### 1. Composable Tutorials with Selected Content Blocks

**Pattern:** A `.. composable-tutorial::` directive wraps a procedure and defines variation options. Within the procedure, `.. selected-content::` blocks provide different content for different selections.

**Example:**

```rst
Connect to Your Cluster
-----------------------

.. composable-tutorial::
   :options: driver, atlas-cli
   :defaults: driver=nodejs; atlas-cli=none

   .. procedure::

      .. step:: Install dependencies

         .. selected-content::
            :selections: driver=nodejs

            Install the MongoDB Node.js driver:

            .. code-block:: bash

               npm install mongodb

         .. selected-content::
            :selections: driver=python

            Install the PyMongo driver:

            .. code-block:: bash

               pip install pymongo

         .. selected-content::
            :selections: atlas-cli=none

            No installation required for Atlas CLI.

      .. step:: Connect to the cluster

         .. selected-content::
            :selections: driver=nodejs

            .. code-block:: javascript

               const { MongoClient } = require('mongodb');
               const client = new MongoClient(uri);

         .. selected-content::
            :selections: driver=python

            .. code-block:: python

               from pymongo import MongoClient
               client = MongoClient(uri)
```

**Parser Behavior:**
- Detects the composable tutorial and extracts options/defaults
- Parses selected-content blocks within steps
- Creates variations: `driver=nodejs`, `driver=python`, `atlas-cli=none`
- **Analysis:** Shows 1 unique procedure with 3 variations
- **Extraction:** Creates 1 file listing all 3 selections

### 2. Tabs Within Steps

**Pattern:** `.. tabs::` directives within procedure steps show different ways to accomplish the same task.

**Example:**

```rst
Procedure with Tabs
-------------------

.. procedure::

   .. step:: Connect to MongoDB

      Choose your programming language:

      .. tabs::

         .. tab:: Node.js
            :tabid: nodejs

            .. code-block:: javascript

               const { MongoClient } = require('mongodb');
               const client = new MongoClient(uri);

         .. tab:: Python
            :tabid: python

            .. code-block:: python

               from pymongo import MongoClient
               client = MongoClient(uri)

         .. tab:: Shell
            :tabid: shell

            .. code-block:: bash

               mongosh "mongodb://localhost:27017"

   .. step:: Verify the connection

      Run a simple query to verify connectivity.
```

**Parser Behavior:**
- Detects tabs within the step content
- Extracts tab IDs: `nodejs`, `python`, `shell`
- Creates variations for each tab
- **Analysis:** Shows 1 unique procedure with 3 variations
- **Extraction:** Creates 1 file listing all 3 tab variations

### 3. Tabs Containing Procedures

**Pattern:** `.. tabs::` directives at the top level contain entirely different procedures for different platforms or contexts.

**Example:**

```rst
Installation Instructions
-------------------------

.. tabs::

   .. tab:: macOS
      :tabid: macos

      .. procedure::

         .. step:: Install Homebrew

            If you don't have Homebrew installed, run:

            .. code-block:: bash

               /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

         .. step:: Install MongoDB

            .. code-block:: bash

               brew tap mongodb/brew
               brew install mongodb-community

         .. step:: Start MongoDB

            .. code-block:: bash

               brew services start mongodb-community

   .. tab:: Ubuntu
      :tabid: ubuntu

      .. procedure::

         .. step:: Import the public key

            .. code-block:: bash

               wget -qO - https://www.mongodb.org/static/pgp/server-6.0.asc | sudo apt-key add -

         .. step:: Create a list file

            .. code-block:: bash

               echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu focal/mongodb-org/6.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-6.0.list

         .. step:: Update package database

            .. code-block:: bash

               sudo apt-get update

         .. step:: Install MongoDB

            .. code-block:: bash

               sudo apt-get install -y mongodb-org

   .. tab:: Windows
      :tabid: windows

      .. procedure::

         .. step:: Download the installer

            Navigate to the MongoDB Download Center and download the Windows installer.

         .. step:: Run the installer

            Double-click the downloaded .msi file and follow the installation wizard.

         .. step:: Configure MongoDB as a service

            During installation, select "Install MongoDB as a Service".
```

**Parser Behavior:**
- Detects tabs at the top level (before any procedure directive)
- Parses each tab's procedure separately
- Each procedure gets a unique content hash (different steps)
- All procedures share a `TabSet` reference for grouping
- **Analysis:** Shows 1 logical procedure with 3 appearances (macos, ubuntu, windows)
- **Extraction:** Creates 3 separate files:
  - `installation-instructions-install-homebrew-87514e.rst` (appears in `macos` selection)
  - `installation-instructions-import-the-public-key-87b686.rst` (appears in `ubuntu` selection)
  - `installation-instructions-download-the-installer-1f0961.rst` (appears in `windows` selection)

**Rationale:** Each platform has a completely different installation procedure with different steps, so they should be extracted as separate files. However, for analysis/reporting, they're grouped as one logical "Installation Instructions" procedure with platform variations.

## Sub-Procedures and List Type Tracking

MongoDB documentation often contains **sub-procedures** - ordered lists within procedure steps that represent nested sequences of actions. The parser tracks these sub-procedures separately and preserves their list marker type (numbered vs. lettered).

### Sub-Procedure Structure

Each step can contain multiple sub-procedures, where each sub-procedure is a separate ordered list:

```rst
.. procedure::

   .. step:: Restart Each Member

      **For each secondary member:**

      a. Shut down the member.

      b. Restart the member.

      **For the primary:**

      a. Step down the primary.

      #. Shut down the member.

      #. Restart the member.
```

In this example, step "Restart Each Member" contains **two separate sub-procedures**:
1. Sub-procedure 1 (2 steps): For each secondary member
2. Sub-procedure 2 (3 steps): For the primary

### List Type Tracking

The parser tracks whether each sub-procedure uses numbered (`1.`, `2.`, `3.`) or lettered (`a.`, `b.`, `c.`) markers:

**Data Structure:**
```go
type SubProcedure struct {
    Steps    []Step // The steps in this sub-procedure
    ListType string // "numbered" or "lettered"
}

type Step struct {
    Title         string
    Content       string
    SubProcedures []SubProcedure // Multiple sub-procedures within this step
}
```

**Parser Behavior:**
- Detects each ordered list within a step as a separate sub-procedure
- Determines list type from the first item (`1.` → numbered, `a.` → lettered)
- Stores each sub-procedure with its list type

### Display with `--show-sub-procedures` Flag

The `extract procedures` command includes a `--show-sub-procedures` flag that displays sub-procedures using their original list marker type:

**Example Output:**
```
Step 2 (Restart Each Member) contains 2 sub-procedures with a total of 5 sub-steps

   Sub-procedure 1 (2 steps):
      a. Shut down the member.
      b. Restart the member.

   Sub-procedure 2 (3 steps):
      a. Step down the primary.
      b. Shut down the member.
      c. Restart the member.
```

**Benefits:**
- Makes it easier for writers to match CLI output with source files
- Preserves the semantic meaning of list marker types
- Shows the structure of multiple sub-procedures within a step

## Include Directive Handling

MongoDB documentation uses `.. include::` directives to reuse content across files. The parser handles includes with context-aware expansion:

### Pattern 1: No Composable Tutorial

If a file has NO composable tutorial, all includes are expanded globally before parsing:

```rst
Simple Procedure
----------------

.. procedure::

   .. step:: First step

      .. include:: /includes/common-setup.rst

   .. step:: Second step

      Do something else.
```

**Parser Behavior:**
- Expands all `.. include::` directives inline
- Then parses the expanded content

### Pattern 2: Composable Tutorial with Selected Content in Main File

If selected-content blocks are in the main file, includes within those blocks are expanded:

```rst
.. composable-tutorial::
   :options: driver
   :defaults: driver=nodejs

   .. procedure::

      .. step:: Install dependencies

         .. selected-content::
            :selections: driver=nodejs

            .. include:: /includes/install-nodejs.rst

         .. selected-content::
            :selections: driver=python

            .. include:: /includes/install-python.rst
```

**Parser Behavior:**
- Detects selected-content blocks
- Expands includes within each block
- Preserves block boundaries

### Pattern 3: Composable Tutorial with Includes Containing Selected Content

If procedure steps include files that contain selected-content blocks:

```rst
.. composable-tutorial::
   :options: driver, atlas-cli
   :defaults: driver=nodejs; atlas-cli=none

   .. procedure::

      .. step:: Install dependencies

         .. include:: /includes/install-deps.rst

      .. step:: Connect to cluster

         .. include:: /includes/connect.rst
```

Where `/includes/install-deps.rst` contains:

```rst
.. selected-content::
   :selections: driver=nodejs

   npm install mongodb

.. selected-content::
   :selections: driver=python

   pip install pymongo
```

**Parser Behavior:**
- When parsing step content, checks if it contains `.. include::` directives
- If no selected-content blocks have been found yet, expands the includes
- Re-parses the expanded content to detect selected-content blocks
- This ensures variations in included files are properly detected

**Rationale:** MongoDB documentation uses composable tutorials inconsistently. Sometimes selected-content blocks are in the main file, sometimes in included files. The parser must handle both patterns.

## Uniqueness and Grouping

The parser uses two mechanisms to identify procedures:

### 1. Heading (Title)

The procedure's heading is the section title above the procedure. For example:

```rst
Connect to Your Cluster
-----------------------

.. procedure::
   ...
```

The heading is "Connect to Your Cluster".

### 2. Content Hash

The content hash is a SHA256 hash of:
- Step titles
- Step content (normalized)
- Variations (sorted for determinism)
- Sub-steps

Two procedures with the same heading but different content will have different hashes.

**Example:**

```rst
# Procedure A
Install MongoDB
---------------
.. procedure::
   .. step:: Download installer
   .. step:: Run installer

# Procedure B
Install MongoDB
---------------
.. procedure::
   .. step:: Install via package manager
   .. step:: Start the service
```

Both have heading "Install MongoDB" but different content hashes because the steps are different.

### Grouping Logic

**For Analysis/Reporting:**
- Procedures are grouped by heading only
- Shows "N unique procedures under this heading"
- Displays all variations for each unique procedure

**For Extraction:**
- Procedures are grouped by heading + content hash
- Each unique procedure (by content hash) is extracted to a separate file
- Filename includes a 6-character hash suffix to prevent collisions

## Analysis vs. Extraction Semantics

The parser has different behavior for analysis vs. extraction:

### Analysis (analyze procedures command)

**Goal:** Give an overview of procedure structure and variations in the documentation.

**Behavior:**
- Groups procedures by heading
- Shows unique procedure count and total appearances
- Lists all variations for each procedure
- **Tabs containing procedures:** Groups all procedures from the same tab set as one logical procedure

**Example Output:**

```
1. Installation Instructions
   Unique procedures: 1
   Total appearances: 3

   Appears in 3 selections:
     - macos
     - ubuntu
     - windows
```

### Extraction (extract procedures command)

**Goal:** Extract each unique procedure to a separate file for reuse.

**Behavior:**
- Creates one file per unique procedure (by content hash)
- Filename includes heading + first step + hash
- Each file lists which selections it appears in
- **Tabs containing procedures:** Creates separate files for each tab's procedure

**Example Output:**

```
Would write: output/installation-instructions-install-homebrew-87514e.rst
  Appears in 1 selection:
    - macos

Would write: output/installation-instructions-import-the-public-key-87b686.rst
  Appears in 1 selection:
    - ubuntu

Would write: output/installation-instructions-download-the-installer-1f0961.rst
  Appears in 1 selection:
    - windows
```

**Rationale:** For analysis, we want to see that there's one logical "Installation Instructions" procedure with platform variations. For extraction, we want separate files because each platform has completely different steps.

## Key Design Decisions

### 1. Deterministic Ordering

**Problem:** Go maps have randomized iteration order, which caused non-deterministic output (procedure counts varied between runs).

**Solution:** All map iterations are sorted by key before processing:
- Tab IDs are sorted alphabetically
- Selected-content selections are sorted
- Variation lists are sorted
- Hash computation uses sorted keys

**Impact:** Ensures consistent output across runs, critical for testing and CI/CD.

### 2. Content Hashing for Uniqueness

**Problem:** Need to detect when two procedures are identical vs. different, even if they have the same heading.

**Solution:** Compute SHA256 hash of normalized step content, including:
- Step titles
- Step content (trimmed, normalized whitespace)
- Variations (sorted)
- Sub-steps (recursively hashed)

**Impact:** Accurately detects duplicate procedures and prevents false grouping.

### 3. Dual-Purpose TabSet Structure

**Problem:** Tabs containing procedures need to be grouped for analysis but extracted separately.

**Solution:**
- Each procedure has a `TabID` field (its specific tab)
- Each procedure has a `TabSet` reference (all procedures in the set)
- `GetProcedureVariations()` returns `TabID` for extraction, `TabSet.TabIDs` for grouping

**Impact:** Same data structure supports both analysis and extraction semantics.

### 4. Context-Aware Include Expansion

**Problem:** Includes can appear at different levels (global, within selected-content, within steps) and need different handling.

**Solution:**
- No composable tutorial → Expand all includes globally
- Composable tutorial with selected-content in main file → Expand includes within blocks
- Composable tutorial with includes in steps → Expand includes to detect selected-content blocks

**Impact:** Handles all patterns of composable tutorial usage in MongoDB docs.

## Common Patterns and Edge Cases

### Pattern: Procedure with No Variations

```rst
Simple Task
-----------

.. procedure::

   .. step:: Do this
   .. step:: Do that
```

**Result:**
- Analysis: 1 unique procedure, 1 appearance, appears in 1 selection (empty string)
- Extraction: 1 file with no selection listed

### Pattern: Multiple Procedures Under Same Heading

```rst
Setup Instructions
------------------

.. procedure::
   .. step:: Install Node.js
   .. step:: Install npm

.. procedure::
   .. step:: Install Python
   .. step:: Install pip
```

**Result:**
- Analysis: 2 unique procedures under "Setup Instructions"
- Extraction: 2 separate files (different content hashes)

### Pattern: Nested Tabs (Tabs Within Tabs)

```rst
.. tabs::
   .. tab:: Platform
      .. tabs::
         .. tab:: macOS
         .. tab:: Windows
```

**Current Behavior:** Only the outer tabs are detected. Inner tabs are treated as regular content.

**Rationale:** Nested tabs are rare in MongoDB docs and add significant complexity. Can be added if needed.

### Pattern: Composable Tutorial with Tabs Within Steps

```rst
.. composable-tutorial::
   :options: driver
   :defaults: driver=nodejs

   .. procedure::
      .. step:: Connect
         .. tabs::
            .. tab:: Async
               :tabid: async
            .. tab:: Sync
               :tabid: sync
```

**Result:**
- Variations are combined: `driver=nodejs; async`, `driver=nodejs; sync`
- Analysis: 1 unique procedure with multiple variations
- Extraction: 1 file listing all combined variations

### Edge Case: Empty Procedure

```rst
.. procedure::
```

**Result:** Skipped (no steps to extract)

### Edge Case: Procedure with Only Sub-steps

```rst
.. procedure::
   .. step:: Main step
      .. procedure::
         .. step:: Sub-step 1
         .. step:: Sub-step 2
```

**Result:**
- Main procedure has 1 step with sub-procedure
- `HasSubSteps` flag is set to true
- Sub-procedure is not extracted separately (only top-level procedures are extracted)

### Pattern: Hierarchical Procedure with Numbered Headings

```rst
Procedure
---------

1. First Major Step
~~~~~~~~~~~~~~~~~~~

Description of the first step.

a. Sub-step one

#. Sub-step two

2. Second Major Step
~~~~~~~~~~~~~~~~~~~~

Description of the second step.

a. Sub-step one

#. Sub-step two
```

**Result:**
- Analysis: 1 unique procedure with 2 steps
- `HasSubSteps` flag is set to true (because of the ordered lists)
- Extraction: 1 file containing both numbered heading steps and their sub-steps

### Pattern: Continuation Markers in Ordered Lists

```rst
Setup Steps
-----------

a. First step

#. Second step (becomes 'b.')

#. Third step (becomes 'c.')
```

**Result:**
- Parser recognizes `#.` as continuation of lettered list
- Converts to: a., b., c.
- Works for both numbered (1., 2., 3.) and lettered (a., b., c.) lists

## Testing Strategy

The parser has comprehensive test coverage:

1. **Unit tests** (`internal/rst/procedure_parser_test.go`):
   - Test each procedure format (directive, ordered list, YAML)
   - Test each variation mechanism (composable tutorials, tabs)
   - Test include expansion
   - Test content hashing determinism

2. **Integration tests** (`commands/analyze/procedures/procedures_test.go`):
   - Test analysis output format
   - Test grouping logic
   - Test deterministic ordering

3. **Extraction tests** (`commands/extract/procedures/procedures_test.go`):
   - Test file generation
   - Test filename uniqueness
   - Test dry-run mode
   - Test selection filtering

4. **Test fixtures** (`testdata/input-files/source/`):
   - `procedure-test.rst`: Comprehensive test file with all patterns
   - `procedure-with-includes.rst`: Tests include expansion
   - `tabs-with-procedures.rst`: Tests tabs containing procedures

## Future Enhancements

Potential improvements to consider:

1. **Nested tabs support**: Handle tabs within tabs
2. **Procedure validation**: Detect malformed procedures and warn users
3. **Cross-file procedure tracking**: Detect when the same procedure appears in multiple files
4. **Variation conflict detection**: Warn when variations have conflicting content
5. **Performance optimization**: Cache parsed procedures for large documentation sets

## Maintenance Guidelines

When modifying the parser:

1. **Update this document** when business logic changes
2. **Update package-level comment** in `procedure_parser.go`
3. **Add test cases** for new patterns or edge cases
4. **Run determinism tests** to ensure consistent output
5. **Check both analysis and extraction** to ensure changes work for both use cases

## Questions?

If you have questions about procedure parsing logic, please:

1. Check this document first
2. Review the package-level comment in `procedure_parser.go`
3. Look at test cases in `*_test.go` files for examples
