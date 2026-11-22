#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-only
#
# Close the macOS-local eBPF loop via Docker (Linux VM + privileged).
# Usage (from repo root):
#   ./scripts/docker-ebpf-verify.sh
#
# Requires: Docker with a Linux engine (Docker Desktop OK).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
IMAGE="${DAE_EBPF_IMAGE:-golang:1.27-bookworm}"
CACHE_ROOT="${DAE_EBPF_CACHE:-/tmp/dae-docker-cache}"
mkdir -p "$CACHE_ROOT/gomod" "$CACHE_ROOT/gocache"

git -C "$ROOT" submodule update --init --recursive

docker run --rm --privileged \
  -v "$ROOT:/build" \
  -v "$CACHE_ROOT/gomod:/go/pkg/mod" \
  -v "$CACHE_ROOT/gocache:/root/.cache/go-build" \
  -w /build \
  -e GOPATH=/go \
  -e GOMODCACHE=/go/pkg/mod \
  -e GOCACHE=/root/.cache/go-build \
  -e GOTOOLCHAIN=local \
  -e GOPROXY="${GOPROXY:-https://proxy.golang.org,direct}" \
  -e CGO_ENABLED=0 \
  "$IMAGE" \
  bash -c '
set -euo pipefail
export PATH=/go/bin:/usr/local/go/bin:$PATH
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq clang llvm make git ca-certificates iproute2 >/dev/null

mkdir -p /sys/kernel/tracing /sys/kernel/debug
mount -t debugfs debugfs /sys/kernel/debug 2>/dev/null || true
mount -t tracefs tracefs /sys/kernel/tracing 2>/dev/null || true
if [[ ! -e /sys/kernel/tracing/trace_pipe && -e /sys/kernel/debug/tracing/trace_pipe ]]; then
  mount --bind /sys/kernel/debug/tracing /sys/kernel/tracing
fi

echo "==> make ebpf"
make ebpf CLANG=clang STRIP=llvm-strip GOFLAGS="-buildvcs=false"
echo "==> make ebpf-test (includes #1110 LAN UDP exact-bind cases)"
make ebpf-test CLANG=clang STRIP=llvm-strip GOFLAGS="-buildvcs=false"

echo "==> production program load / ABI whitelist"
go test -c ./control -o /tmp/dae-control-datapath.test
whitelist=(
  TestOffloadObjectsInSpec
  TestDataplaneProgramsAndMapsComplete
  TestAssignDataplaneToBpfCopiesMandatoryObjects
  TestBatchDeleteIgnoringMissingResumesSuffix
  TestBatchDeleteIgnoringMissingAllAbsent
  TestBatchDeleteIgnoringMissingPreservesPartialCountOnError
  TestBatchDeleteIgnoringMissingRejectsSilentPartialProgress
  TestBpfMapBatchDeleteRealMapResumesAfterMissingKey
  TestRemovedIncompatiblePinnedMapWarnsWithConsequence
  TestIncompatiblePinRemovalIsNotAnInfoStep
  TestDetectCgroupPathCaching
  TestProductionProgramsLoad
  TestDaeParamLayout
  TestDirectStatsConnStateABIParity
  TestDirectStatsTuplesKeyABIParity
  TestDirectStatsObjectHasNoNATReverseMap
  TestDirectStatsCounterValueLayout
  TestDirectStatsProgramsLoad
)
# TestTCHookSetKernelUpdateAndRollback/continuation needs a real TC packet
# path; Docker Desktop linuxkit often drops it. Run it soft: report, do not fail.
run_regex="^($(IFS=\|; echo "${whitelist[*]}"))$"
/tmp/dae-control-datapath.test -test.run "$run_regex" -test.count=1 -test.v | tee /tmp/datapath.log
failed=0
for name in "${whitelist[@]}"; do
  if ! grep -qE -- "--- PASS: ${name}([[:space:]]|$)" /tmp/datapath.log; then
    echo "FAIL: $name did not PASS"
    failed=1
  fi
done
[[ "$failed" -eq 0 ]]

echo "==> soft: TC hook set (may fail on Docker Desktop)"
/tmp/dae-control-datapath.test -test.run "^TestTCHookSetKernelUpdateAndRollback$" -test.count=1 -test.v || \
  echo "WARN: TestTCHookSetKernelUpdateAndRollback soft-failed (linuxkit TC packet path)"

TAGS=$(tr -d "\n" < .build_tags || true)
echo "==> go build tags=${TAGS:-<empty>}"
GOOS=linux go build -tags="$TAGS" -o /tmp/dae-linux -trimpath .
ls -la /tmp/dae-linux
echo "DOCKER_EBPF_VERIFY_OK"
'
