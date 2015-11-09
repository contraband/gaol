#!/usr/bin/env bats

load test_helper

@test "a created container appears in the list" {
  handle=$(gaol create)

  run gaol list

  assert_success
  assert_match $handle
}

@test "a created container can have its handle chosen" {
  run gaol create -n awesome-handle

  assert_success
  assert_match $handle

  run gaol list

  assert_success
  assert_match $handle
}

