#!/usr/bin/env python3
"""
Test Pattern Matching Against Real Files
This script helps you understand which files will match your copier patterns
"""

import re
from fnmatch import fnmatch

# Test files (common files in mflix directory)
TEST_FILES = [
    "mflix/client/src/App.tsx",
    "mflix/client/src/components/Header.tsx",
    "mflix/client/package.json",
    "mflix/client/.gitignore",
    "mflix/client/README.md",
    "mflix/server/java-spring/src/main/java/Main.java",
    "mflix/server/java-spring/pom.xml",
    "mflix/server/java-spring/.env",
    "mflix/server/js-express/src/index.js",
    "mflix/server/js-express/package.json",
    "mflix/server/python-fastapi/main.py",
    "mflix/server/python-fastapi/requirements.txt",
    "mflix/README-JAVA-SPRING.md",
    "mflix/README-JAVASCRIPT-EXPRESS.md",
    "mflix/README-PYTHON-FASTAPI.md",
    "mflix/.gitignore-java",
    "mflix/.gitignore-js",
    "mflix/.gitignore-python",
    "mflix/docker-compose.yml",
    "mflix/package.json",
    "mflix/README.md",
    "other/file.txt",
]

# Pattern definitions from your config
PATTERNS = {
    "mflix-client-to-java": ("prefix", "mflix/client/"),
    "java-server": ("regex", r"^mflix/server/java-spring/(?P<file>.+)$"),
    "mflix-java-readme": ("glob", "mflix/README-JAVA-SPRING.md"),
    "mflix-java-gitignore": ("glob", "mflix/.gitignore-java"),
    "mflix-client-to-js": ("prefix", "mflix/client/"),
    "mflix-express-server": ("regex", r"^mflix/server/js-express/(?P<file>.+)$"),
    "mflix-js-readme": ("glob", "mflix/README-JAVASCRIPT-EXPRESS.md"),
    "mflix-js-gitignore": ("glob", "mflix/.gitignore-js"),
    "mflix-client-to-python": ("prefix", "mflix/client/"),
    "mflix-python-server": ("regex", r"^mflix/server/python-fastapi/(?P<file>.+)$"),
    "mflix-python-readme": ("glob", "mflix/README-PYTHON-FASTAPI.md"),
    "mflix-python-gitignore": ("glob", "mflix/.gitignore-python"),
}

# Exclusion patterns
EXCLUSIONS = [
    r"\.gitignore$",
    r"README\.md$",
    r"\.env$",
]


def is_excluded(file_path):
    """Check if file matches any exclusion pattern"""
    for pattern in EXCLUSIONS:
        if re.search(pattern, file_path):
            return True
    return False


def test_prefix(file_path, pattern):
    """Test if file matches prefix pattern"""
    return file_path.startswith(pattern)


def test_regex(file_path, pattern):
    """Test if file matches regex pattern"""
    return re.match(pattern, file_path) is not None


def test_glob(file_path, pattern):
    """Test if file matches glob pattern"""
    return fnmatch(file_path, pattern)


def test_file(file_path):
    """Test a file against all patterns"""
    # Check if excluded first
    if is_excluded(file_path):
        return "excluded", []
    
    # Test against each pattern
    matching_rules = []
    for rule_name, (pattern_type, pattern) in PATTERNS.items():
        matched = False
        if pattern_type == "prefix":
            matched = test_prefix(file_path, pattern)
        elif pattern_type == "regex":
            matched = test_regex(file_path, pattern)
        elif pattern_type == "glob":
            matched = test_glob(file_path, pattern)
        
        if matched:
            matching_rules.append(rule_name)
    
    if matching_rules:
        return "matched", matching_rules
    else:
        return "skipped", []


def main():
    print("=" * 60)
    print("Pattern Matching Test Tool")
    print("=" * 60)
    print()
    
    matched_count = 0
    skipped_count = 0
    excluded_count = 0
    skipped_files = []
    
    for file_path in TEST_FILES:
        status, rules = test_file(file_path)
        
        if status == "excluded":
            print(f"🟡 EXCLUDED  {file_path}")
            print(f"             └─ Matches exclusion pattern (by design)")
            excluded_count += 1
        elif status == "matched":
            print(f"✅ MATCHED   {file_path}")
            for rule in rules:
                print(f"             └─ Rule: {rule}")
            matched_count += 1
        else:  # skipped
            print(f"❌ SKIPPED   {file_path}")
            print(f"             └─ No matching rules!")
            skipped_count += 1
            skipped_files.append(file_path)
        print()
    
    # Summary
    print("=" * 60)
    print("Summary")
    print("=" * 60)
    print(f"Total files:     {len(TEST_FILES)}")
    print(f"✅ Matched:      {matched_count}")
    print(f"🟡 Excluded:     {excluded_count} (by design)")
    print(f"❌ Skipped:      {skipped_count} (PROBLEM!)")
    print()
    
    if skipped_count > 0:
        print(f"⚠️  WARNING: {skipped_count} files will NOT be copied!")
        print()
        print("These files don't match any pattern in your config:")
        for file_path in skipped_files:
            print(f"  - {file_path}")
        print()
        print("If they should be copied, you need to add rules for them.")
        print()
        print("Common fixes:")
        print("  1. Add a catch-all rule for mflix/ directory")
        print("  2. Add specific rules for these file types")
        print("  3. Verify these files should actually be excluded")
    else:
        print("✅ All non-excluded files have matching rules!")
    
    print()
    print("=" * 60)
    print("Next Steps")
    print("=" * 60)
    print("1. Review the SKIPPED files above")
    print("2. Decide if they should be copied")
    print("3. If yes, add patterns to copier-config.yaml")
    print("4. Run this script again to verify")
    print()
    print("To test against REAL files from your repo:")
    print("  cd /path/to/docs-sample-apps")
    print("  git diff --name-only main")
    print()


if __name__ == "__main__":
    main()

