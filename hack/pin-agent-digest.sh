#!/usr/bin/env bash
set -euo pipefail

# pin-agent-digest.sh <container_tool> <agent_img_versioned> <agent_img_latest> <override_digest1> <override_digest2>
# Adds/updates agent image digest reference (AGENT_IMAGE env) + relatedImages entry, ensures minKubeVersion,
# adds description/displayName for IPRuleConfig and injects an IPRuleConfig example into alm-examples if missing.

CTOOL=${1:-podman}
IMG_VERSION=${2:-}
IMG_LATEST=${3:-}
OVERRIDE1=${4:-}
OVERRIDE2=${5:-}

AGENT_REPO="ghcr.io/mariusbertram/iprule-agent"
CSV_FILE=$(ls bundle/manifests/*clusterserviceversion.yaml 2>/dev/null || true)
if [[ -z "${CSV_FILE}" ]]; then
  echo "[pin-agent-digest] No CSV file found (run make bundle first)." >&2
  exit 0
fi

resolve_digest() {
  local d=""
  local img="$1"
  [[ -z "${img}" ]] && return 0
  ${CTOOL} pull "${img}" >/dev/null 2>&1 || true
  d=$(${CTOOL} image inspect "${img}" --format '{{index .RepoDigests 0}}' 2>/dev/null | head -n1 || true)
  if [[ -z "${d}" ]]; then
    d=$(${CTOOL} image inspect "${img}" --format '{{.Digest}}' 2>/dev/null | head -n1 || true)
  fi
  case "${d}" in
    *@sha256:*) echo "${d#*@}" ;;
    sha256:*) echo "${d}" ;;
  esac
}

DIGEST=""
if [[ -n "${OVERRIDE1}" ]]; then
  DIGEST=${OVERRIDE1}
elif [[ -n "${OVERRIDE2}" ]]; then
  DIGEST=${OVERRIDE2}
else
  DIGEST=$(resolve_digest "${IMG_VERSION}")
  [[ -z "${DIGEST}" ]] && DIGEST=$(resolve_digest "${IMG_LATEST}") || true
fi
# Validate format
if [[ -n "${DIGEST}" && ${DIGEST} != sha256:* ]]; then
  echo "[pin-agent-digest] WARN: Unexpected digest format: ${DIGEST}" >&2
  DIGEST=""
fi
REF=""
if [[ -n "${DIGEST}" ]]; then
  REF="${AGENT_REPO}@${DIGEST}"
fi

# Always ensure AGENT_IMAGE env is set to REF if REF present
if [[ -n "${REF}" ]]; then
  # Replace either :latest or any existing @sha256:... reference
  if grep -q 'name: AGENT_IMAGE' "${CSV_FILE}"; then
    sed -i -E "0,/value: ${AGENT_REPO}(@sha256:[0-9a-f]{64}|:latest)?/ s|value: ${AGENT_REPO}(@sha256:[0-9a-f]{64}|:latest)?|value: ${REF}|" "${CSV_FILE}" || true
  fi
fi

# Ensure relatedImages contains agent REF (add or replace)
if [[ -n "${REF}" ]]; then
  if grep -q '^[[:space:]]*relatedImages:' "${CSV_FILE}"; then
    if grep -q 'name: agent' "${CSV_FILE}"; then
      # replace existing agent line(s)
      awk -v REF="${REF}" 'BEGIN{pending=0} {
        if($0 ~ /name: agent/){print "  - image: " REF "\n    name: agent"; pending=1; next}
        if(pending){pending=0; next}
        print
      }' "${CSV_FILE}" > "${CSV_FILE}.tmp" && mv "${CSV_FILE}.tmp" "${CSV_FILE}"
    else
      # append under relatedImages block (after last existing relatedImages entry)
      awk -v REF="${REF}" 'BEGIN{ri=0} {
        print $0; if($0 ~ /^[[:space:]]*relatedImages:/){ri=1; next}
        if(ri && $0 !~ /^  - /){print "  - image: " REF "\n    name: agent"; ri=0}
      } END{if(ri){print "  - image: " REF "\n    name: agent"}}' "${CSV_FILE}" > "${CSV_FILE}.tmp" && mv "${CSV_FILE}.tmp" "${CSV_FILE}"
    fi
  else
    printf "  relatedImages:\n  - image: %s\n    name: agent\n" "${REF}" >> "${CSV_FILE}"
  fi
fi

# Add minKubeVersion if missing
if ! grep -q '^  minKubeVersion:' "${CSV_FILE}"; then
  sed -i '/^  version: /a \\  minKubeVersion: "1.28.0"' "${CSV_FILE}" || true
fi

# Add description/displayName for IPRuleConfig if missing
if grep -q 'kind: IPRuleConfig' "${CSV_FILE}"; then
  if ! awk '/kind: IPRuleConfig/{found=1} found && /displayName: IPRuleConfig/{ok=1} END{exit ok}' "${CSV_FILE}"; then
    # Insert two lines before first - kind: IPRuleConfig
    sed -i '/^- kind: IPRuleConfig/i \
      description: IPRuleConfig represents a desired IP rule configuration for a Service IP\n      displayName: IPRuleConfig' "${CSV_FILE}" || true
  fi
fi

# Inject IPRuleConfig example into alm-examples JSON if absent
if ! awk '/alm-examples:/{in=1} in && /"kind": *"IPRuleConfig"/{found=1} in && /\]/{exit} END{exit found}' "${CSV_FILE}"; then
  awk 'BEGIN{in=0;done=0} /alm-examples:/{in=1} {
    if(in && !done && /\]/){
      print "        ,";
      print "        {";
      print "          \"apiVersion\": \"api.operator.brtrm.dev/v1alpha1\",";
      print "          \"kind\": \"IPRuleConfig\",";
      print "          \"metadata\": {\"name\": \"ipruleconfig-sample\"},";
      print "          \"spec\": {\"serviceIP\": \"10.0.0.10\", \"table\": 254, \"priority\": 100, \"state\": \"Present\"}";
      print "        }";
      done=1;
    }
    print $0
  }' "${CSV_FILE}" > "${CSV_FILE}.tmp" && mv "${CSV_FILE}.tmp" "${CSV_FILE}"
fi

# Report
if [[ -n "${REF}" ]]; then
  if grep -q "${REF}" "${CSV_FILE}"; then
    echo "[pin-agent-digest] Agent image pinned to ${REF}";
  else
    echo "[pin-agent-digest] FAILED to pin agent image" >&2
  fi
else
  echo "[pin-agent-digest] No digest pinned (image not built/pushed or no override)." >&2
fi

grep -n 'minKubeVersion:' "${CSV_FILE}" || true
awk '/^- kind: IPRuleConfig/{print NR":"$0;exit}' "${CSV_FILE}" || true
awk '/alm-examples:/{in=1} in && /"kind": *"IPRuleConfig"/{print NR":"$0; exit}' "${CSV_FILE}" || true
