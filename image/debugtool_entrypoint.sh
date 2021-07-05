#!/usr/bin/env bash
set -xeuo

select_ip_by_cidr() {
  cidr=$1
  cidr_network=$(ipcalc -n "$cidr" | cut -f2 -d=)
  cidr_prefix=$(ipcalc -p "$cidr" | cut -f2 -d=)

  IFS=' ' read -r -a ips <<<"$(echo $(ip addr show | grep inet | grep -v inet6 | awk '{print $2}'))"
  results=()
  for ip in ${ips[@]}; do
    ip_network="$(ipcalc -n $(echo ${ip} | awk -F'/' '{print $1}')/${cidr_prefix} | cut -f2 -d=)"

    if [ "$ip_network" = "$cidr_network" ]; then
      results+=("$(echo ${ip} | awk -F'/' '{print $1}')")
    fi
  done
  if [ "${#results[@]}" -gt 1 ]; then
    echo "[ERROR] get more than 1 ip with cidr ${cidr}" >&2
    return
  fi
  echo "${results[0]}"
}

if [[ -z $HOST_NETWORK_CIDR ]]; then
	IPERF_BIND_ADDR="0.0.0.0"
else
	IPERF_BIND_ADDR=$(select_ip_by_cidr "$HOST_NETWORK_CIDR")
	if [[ -z "$IPERF_BIND_ADDR" ]]; then
	   echo "warning IPERF_BIND_ADDR is empty"
	fi
fi

echo $IPERF_BIND_ADDR > /opt/iperf_bind_addr

export IPERF_BIND_ADDR

iperf3 -s -B $IPERF_BIND_ADDR
