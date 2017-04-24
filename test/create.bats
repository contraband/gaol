#!/usr/bin/env bats

load test_helper

@test "a created container appears in the list" {
  handle=$(gaol create)

  run gaol list

  assert_success
  assert_match $handle
}

@test "a container can be created with memory limits" {
  handle=$(gaol create --limit-memory 52428800)
  assert_success

  run gaol properties $handle
  assert_success
  assert_match "memory.limit\t52428800 bytes"
}

@test "a created container can have its handle chosen" {
  run gaol create -n awesome-handle

  assert_success
  assert_match $handle

  run gaol list

  assert_success
  assert_match $handle
}

