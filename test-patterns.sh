#!/bin/bash
# Test Pattern Matching Against Real Files
# This script helps you understand which files will match your copier patterns

set -e

echo "========================================="
echo "Pattern Matching Test Tool"
echo "========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test files (common files in mflix directory)
TEST_FILES=(
    "mflix/client/src/App.tsx"
    "mflix/client/src/components/Header.tsx"
    "mflix/client/package.json"
    "mflix/client/.gitignore"
    "mflix/client/README.md"
    "mflix/server/java-spring/src/main/java/Main.java"
    "mflix/server/java-spring/pom.xml"
    "mflix/server/java-spring/.env"
    "mflix/server/js-express/src/index.js"
    "mflix/server/js-express/package.json"
    "mflix/server/python-fastapi/main.py"
    "mflix/server/python-fastapi/requirements.txt"
    "mflix/README-JAVA-SPRING.md"
    "mflix/README-JAVASCRIPT-EXPRESS.md"
    "mflix/README-PYTHON-FASTAPI.md"
    "mflix/.gitignore-java"
    "mflix/.gitignore-js"
    "mflix/.gitignore-python"
    "mflix/docker-compose.yml"
    "mflix/package.json"
    "mflix/README.md"
    "other/file.txt"
)

# Pattern definitions from your config
declare -A PATTERNS
PATTERNS["mflix-client-to-java"]="prefix:mflix/client/"
PATTERNS["java-server"]="regex:^mflix/server/java-spring/(?P<file>.+)$"
PATTERNS["mflix-java-readme"]="glob:mflix/README-JAVA-SPRING.md"
PATTERNS["mflix-java-gitignore"]="glob:mflix/.gitignore-java"
PATTERNS["mflix-client-to-js"]="prefix:mflix/client/"
PATTERNS["mflix-express-server"]="regex:^mflix/server/js-express/(?P<file>.+)$"
PATTERNS["mflix-js-readme"]="glob:mflix/README-JAVASCRIPT-EXPRESS.md"
PATTERNS["mflix-js-gitignore"]="glob:mflix/.gitignore-js"
PATTERNS["mflix-client-to-python"]="prefix:mflix/client/"
PATTERNS["mflix-python-server"]="regex:^mflix/server/python-fastapi/(?P<file>.+)$"
PATTERNS["mflix-python-readme"]="glob:mflix/README-PYTHON-FASTAPI.md"
PATTERNS["mflix-python-gitignore"]="glob:mflix/.gitignore-python"

# Exclusion patterns
EXCLUSIONS=(
    "\\.gitignore$"
    "README.md$"
    "\\.env$"
)

# Function to check if file matches exclusion
is_excluded() {
    local file=$1
    for pattern in "${EXCLUSIONS[@]}"; do
        if echo "$file" | grep -qE "$pattern"; then
            return 0  # true, is excluded
        fi
    done
    return 1  # false, not excluded
}

# Function to test prefix pattern
test_prefix() {
    local file=$1
    local pattern=$2
    if [[ "$file" == "$pattern"* ]]; then
        return 0  # match
    fi
    return 1  # no match
}

# Function to test regex pattern
test_regex() {
    local file=$1
    local pattern=$2
    # Remove named groups for bash regex
    local bash_pattern=$(echo "$pattern" | sed 's/(?P<[^>]*>/(/g')
    if [[ "$file" =~ $bash_pattern ]]; then
        return 0  # match
    fi
    return 1  # no match
}

# Function to test glob pattern
test_glob() {
    local file=$1
    local pattern=$2
    if [[ "$file" == $pattern ]]; then
        return 0  # match
    fi
    return 1  # no match
}

# Test each file
echo "Testing ${#TEST_FILES[@]} files against ${#PATTERNS[@]} patterns..."
echo ""

matched_count=0
skipped_count=0
excluded_count=0

for file in "${TEST_FILES[@]}"; do
    matched=false
    excluded=false
    matching_rules=()
    
    # Check if excluded first
    if is_excluded "$file"; then
        excluded=true
        excluded_count=$((excluded_count + 1))
    fi
    
    # Test against each pattern
    for rule_name in "${!PATTERNS[@]}"; do
        pattern_def="${PATTERNS[$rule_name]}"
        pattern_type="${pattern_def%%:*}"
        pattern="${pattern_def#*:}"
        
        case "$pattern_type" in
            prefix)
                if test_prefix "$file" "$pattern"; then
                    matching_rules+=("$rule_name")
                    matched=true
                fi
                ;;
            regex)
                if test_regex "$file" "$pattern"; then
                    matching_rules+=("$rule_name")
                    matched=true
                fi
                ;;
            glob)
                if test_glob "$file" "$pattern"; then
                    matching_rules+=("$rule_name")
                    matched=true
                fi
                ;;
        esac
    done
    
    # Print result
    if [ "$excluded" = true ]; then
        echo -e "${YELLOW}EXCLUDED${NC} $file"
        echo "         └─ Matches exclusion pattern (by design)"
    elif [ "$matched" = true ]; then
        echo -e "${GREEN}MATCHED${NC}  $file"
        for rule in "${matching_rules[@]}"; do
            echo "         └─ Rule: $rule"
        done
        matched_count=$((matched_count + 1))
    else
        echo -e "${RED}SKIPPED${NC}  $file"
        echo "         └─ No matching rules!"
        skipped_count=$((skipped_count + 1))
    fi
    echo ""
done

# Summary
echo "========================================="
echo "Summary"
echo "========================================="
echo -e "Total files:     ${#TEST_FILES[@]}"
echo -e "${GREEN}Matched:${NC}         $matched_count"
echo -e "${YELLOW}Excluded:${NC}        $excluded_count (by design)"
echo -e "${RED}Skipped:${NC}         $skipped_count (PROBLEM!)"
echo ""

if [ $skipped_count -gt 0 ]; then
    echo -e "${RED}⚠️  WARNING: $skipped_count files will NOT be copied!${NC}"
    echo ""
    echo "These files don't match any pattern in your config."
    echo "If they should be copied, you need to add rules for them."
    echo ""
    echo "Common fixes:"
    echo "  1. Add a catch-all rule for mflix/ directory"
    echo "  2. Add specific rules for these file types"
    echo "  3. Verify these files should actually be excluded"
else
    echo -e "${GREEN}✅ All non-excluded files have matching rules!${NC}"
fi

echo ""
echo "========================================="
echo "Next Steps"
echo "========================================="
echo "1. Review the SKIPPED files above"
echo "2. Decide if they should be copied"
echo "3. If yes, add patterns to copier-config.yaml"
echo "4. Run this script again to verify"
echo ""
echo "To test against REAL files from your repo:"
echo "  cd /path/to/docs-sample-apps"
echo "  git diff --name-only main | while read file; do"
echo "    echo \"Testing: \$file\""
echo "  done"
echo ""

