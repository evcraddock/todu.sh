#!/bin/bash
# Test script to verify todu-api active tasks with due dates

API_URL="http://10.10.1.151:8000"

curl -s "http://10.10.1.151:8000/api/v1/tasks/?status=active&limit=500" | jq '{total: .total, returned: (.items | length), has_206: ([.items[] | select(.id == 206)] | length > 0), has_207: ([.items[] | select(.id == 207)] | length > 0)}'
