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

@test "a created container bind mount is read-write by default" {
  handle=$(gaol create -m .:/tmp/cmnt)
  assert_success

  run gaol run -a -c "/bin/sh -c 'cat /proc/self/mounts | grep /tmp/cmnt'" $handle

  assert_success
  assert_match " rw,"
}

@test "a created container can have read-only bind mounts" {
  handle=$(gaol create -m .:/tmp/cmnt:ro)
  assert_success

  run gaol run -a -c "/bin/sh -c 'cat /proc/self/mounts | grep /tmp/cmnt'" $handle

  assert_success
  assert_match " ro,"
}


@test "a created container can have explicit read-write bind mounts" {
  handle=$(gaol create -m .:/tmp/cmnt:rw)
  assert_success

  run gaol run -a -c "/bin/sh -c 'cat /proc/self/mounts | grep /tmp/cmnt'" $handle

  assert_success
  assert_match " rw,"
}

@test "creation fails if unsupported bind mount mode is provided" {
  run gaol create -m .:/tmp/cmnt:foo
  assert_failure
}
