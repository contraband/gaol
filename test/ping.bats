#!/usr/bin/env bats

load test_helper

@test "pinging an active garden server succeeds" {
  run gaol ping

  assert_success
}

@test "pinging an inactive garden server succeeds" {
  run gaol -t garden.example.com:7777 ping

  assert_failure
}
