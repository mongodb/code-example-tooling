#!/usr/bin/env python3
"""
Detailed validation script for copier-config.yaml files
"""

import sys
import yaml
import re

def validate_config(file_path):
    """Validate a copier-config.yaml file and report all issues"""
    
    issues = []
    warnings = []
    
    # Try to load the YAML
    try:
        with open(file_path, 'r') as f:
            config = yaml.safe_load(f)
    except yaml.YAMLError as e:
        print(f"❌ YAML Parsing Error:")
        print(f"   {e}")
        return False
    except Exception as e:
        print(f"❌ Error reading file: {e}")
        return False
    
    print("✅ YAML syntax is valid")
    print()
    
    # Validate structure
    if not isinstance(config, dict):
        issues.append("Config must be a dictionary")
        return False
    
    # Check required fields
    if 'source_repo' not in config:
        issues.append("Missing required field: source_repo")
    
    if 'copy_rules' not in config:
        issues.append("Missing required field: copy_rules")
    
    if issues:
        print("❌ Structural Issues:")
        for issue in issues:
            print(f"   - {issue}")
        return False
    
    print(f"📋 Config Summary:")
    print(f"   Source: {config.get('source_repo')}")
    print(f"   Branch: {config.get('source_branch', 'main')}")
    print(f"   Rules: {len(config.get('copy_rules', []))}")
    print()
    
    # Validate each rule
    rules = config.get('copy_rules', [])
    for i, rule in enumerate(rules, 1):
        rule_name = rule.get('name', f'Rule {i}')
        print(f"🔍 Validating Rule {i}: {rule_name}")
        
        # Check rule structure
        if 'source_pattern' not in rule:
            issues.append(f"Rule '{rule_name}': Missing source_pattern")
            continue
        
        if 'targets' not in rule:
            issues.append(f"Rule '{rule_name}': Missing targets")
            continue
        
        # Validate source_pattern
        pattern = rule['source_pattern']
        if not isinstance(pattern, dict):
            issues.append(f"Rule '{rule_name}': source_pattern must be a dictionary")
            continue
        
        pattern_type = pattern.get('type')
        pattern_str = pattern.get('pattern')
        
        if not pattern_type:
            issues.append(f"Rule '{rule_name}': Missing pattern type")
        elif pattern_type not in ['prefix', 'glob', 'regex']:
            issues.append(f"Rule '{rule_name}': Invalid pattern type '{pattern_type}' (must be prefix, glob, or regex)")
        
        if not pattern_str:
            issues.append(f"Rule '{rule_name}': Missing pattern string")
        else:
            # Check for type/pattern mismatch
            has_regex_syntax = bool(re.search(r'\(\?P<\w+>', pattern_str))
            
            if pattern_type == 'prefix' and has_regex_syntax:
                issues.append(f"Rule '{rule_name}': Pattern type is 'prefix' but pattern contains regex syntax '(?P<...>)'")
                warnings.append(f"Rule '{rule_name}': Should use type: 'regex' instead of 'prefix'")
            
            # Validate regex patterns
            if pattern_type == 'regex':
                try:
                    re.compile(pattern_str)
                except re.error as e:
                    issues.append(f"Rule '{rule_name}': Invalid regex pattern: {e}")
        
        # Validate targets
        targets = rule.get('targets', [])
        if not isinstance(targets, list):
            issues.append(f"Rule '{rule_name}': targets must be a list")
            continue
        
        if len(targets) == 0:
            warnings.append(f"Rule '{rule_name}': No targets defined")
        
        for j, target in enumerate(targets, 1):
            if not isinstance(target, dict):
                issues.append(f"Rule '{rule_name}', Target {j}: Must be a dictionary")
                continue
            
            # Check required target fields
            if 'repo' not in target:
                issues.append(f"Rule '{rule_name}', Target {j}: Missing 'repo' field")
            
            if 'branch' not in target:
                warnings.append(f"Rule '{rule_name}', Target {j}: Missing 'branch' field (will use default)")
            
            if 'path_transform' not in target:
                warnings.append(f"Rule '{rule_name}', Target {j}: Missing 'path_transform' field")
            
            # Validate commit_strategy
            if 'commit_strategy' in target:
                strategy = target['commit_strategy']
                if not isinstance(strategy, dict):
                    issues.append(f"Rule '{rule_name}', Target {j}: commit_strategy must be a dictionary")
                else:
                    strategy_type = strategy.get('type')
                    if strategy_type and strategy_type not in ['direct', 'pull_request']:
                        issues.append(f"Rule '{rule_name}', Target {j}: Invalid commit_strategy type '{strategy_type}'")
        
        print(f"   ✓ Rule validated")
    
    print()
    
    # Print summary
    if issues:
        print("❌ VALIDATION FAILED")
        print()
        print("Issues found:")
        for issue in issues:
            print(f"   ❌ {issue}")
        print()
    
    if warnings:
        print("⚠️  Warnings:")
        for warning in warnings:
            print(f"   ⚠️  {warning}")
        print()
    
    if not issues and not warnings:
        print("✅ Configuration is valid with no issues!")
        return True
    elif not issues:
        print("✅ Configuration is valid (with warnings)")
        return True
    else:
        return False

if __name__ == '__main__':
    if len(sys.argv) != 2:
        print("Usage: validate-config-detailed.py <config-file>")
        sys.exit(1)
    
    file_path = sys.argv[1]
    success = validate_config(file_path)
    sys.exit(0 if success else 1)

