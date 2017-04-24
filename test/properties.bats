#!/usr/bin/env bats

load test_helper

@test "a container reports its limits" {
  run gaol properties $(gaol create)
  assert_success
  assert_match "memory.limit\t0 bytes"
}
