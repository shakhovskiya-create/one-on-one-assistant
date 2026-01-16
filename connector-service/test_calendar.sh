#!/bin/bash
# Test EWS calendar access

EMAIL="a.shakhovskiy@ekf.su"
echo "Testing calendar access for: $EMAIL"
echo "---"

# Try to get calendar via connector (if it has a direct test endpoint)
# For now, let's just check if connector can reach EWS

curl -s -k https://post.ekf.su/EWS/Exchange.asmx | head -20
