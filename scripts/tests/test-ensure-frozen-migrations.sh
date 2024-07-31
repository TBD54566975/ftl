#!/bin/bash
set -euo pipefail
set -x

script_path='scripts/ensure-frozen-migrations'
schema_dir='backend/controller/sql/schema'

function create() {
  echo "CREATE TABLE testing (id SERIAL PRIMARY KEY);" > "$schema_dir/testing.sql"
}

function remove_existing() {
  rm "$schema_dir/20240704103403_create_module_secrets.sql"
}

function modify_comment_in_existing() {
  sed -i.bak 's/Function for deployment notifications./🧐/' "$schema_dir/20231103205514_init.sql"
  rm "$schema_dir/20231103205514_init.sql.bak"
}

function modify_content_in_existing() {
  sed -i.bak 's/CREATE OR REPLACE FUNCTION notify_event()/CREATE 🥲 notify_event()/' "$schema_dir/20231103205514_init.sql"
  rm "$schema_dir/20231103205514_init.sql.bak"
}

function no_changes() {
  echo "no-changes" >> .gitignore
  git add .gitignore
}

function commit_migrations_dir() {
  git add "$schema_dir"
  git commit -m "ci: automated test commit, this should not be pushed!"
}


# higher order test function, accepting a function as an argument, and an expected fail or pass.
# omitted means it should pass.
# fail should be used when the test is expected to fail.
function test() {
  local test_function=$1
  local expected_result=${2:-pass}
  local saved_commit
  saved_commit=$(git rev-parse HEAD)

  echo "---- $test_function -------------------------"

  $test_function

  git status
  commit_migrations_dir

  local did_fail=false
  if ! $script_path; then
    if [ "$expected_result" == "pass" ]; then
      echo "❌ $test_function expected to pass but failed"
      did_fail=true
    fi
  else
    if [ "$expected_result" == "fail" ]; then
      echo "❌ $test_function expected to fail but passed"
      did_fail=true
    fi
  fi

  # only clean up the schema dir
  rm "$schema_dir"/*
  git reset "$saved_commit"
  git checkout .gitignore "$schema_dir"

  if $did_fail; then
    echo "❌ FAIL $test_function"
    exit 1
  else
    echo "✅ PASS $test_function"
  fi
}

function main() {
  # cd into the project root
  cd "$(git rev-parse --show-toplevel)" || exit 1

  test no_changes
  test create
  test remove_existing fail
  test modify_comment_in_existing
  test modify_content_in_existing fail

  echo "✅ All tests passed 🎉"
}

main