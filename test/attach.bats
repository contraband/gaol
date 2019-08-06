#!/usr/bin/env bats

load test_helper

@test "it returns exit code from attach" {
  handle=$(gaol create)
  assert_success

  process_id=$(gaol run -c 'sh -c "sleep 1; exit 42"' $handle)
  assert_success

  run gaol attach $handle -p $process_id
  assert_equal $status 42
}
