#!/bin/bash

# Running commands and capturing output
echo "[COMMAND 1]"
echo "backlogit query \"SELECT id,artifact_type,status,parent_id FROM items WHERE id LIKE '019%' OR id LIKE '020%' OR id LIKE '021%' OR id LIKE '023%' OR id LIKE '024%' OR id LIKE '025%' ORDER BY id\""
echo "[OUTPUT]"
backlogit query "SELECT id,artifact_type,status,parent_id FROM items WHERE id LIKE '019%' OR id LIKE '020%' OR id LIKE '021%' OR id LIKE '023%' OR id LIKE '024%' OR id LIKE '025%' ORDER BY id" 2>&1
echo ""

echo "[COMMAND 2]"
echo "backlogit get 025-F"
echo "[OUTPUT]"
backlogit get 025-F 2>&1
echo ""

echo "[COMMAND 3]"
echo "backlogit get 025.001-T"
echo "[OUTPUT]"
backlogit get 025.001-T 2>&1
echo ""

echo "[COMMAND 4]"
echo "backlogit dep list 025.001-T"
echo "[OUTPUT]"
backlogit dep list 025.001-T 2>&1
echo ""

echo "[COMMAND 5]"
echo "backlogit query \"SELECT item_id,depends_on,dep_type FROM item_deps WHERE depends_on='025.001-T'\""
echo "[OUTPUT]"
backlogit query "SELECT item_id,depends_on,dep_type FROM item_deps WHERE depends_on='025.001-T'" 2>&1
echo ""

echo "[COMMAND 6]"
echo "backlogit query \"SELECT item_id,depends_on,dep_type FROM item_deps WHERE item_id LIKE '019%' OR item_id LIKE '025%'\""
echo "[OUTPUT]"
backlogit query "SELECT item_id,depends_on,dep_type FROM item_deps WHERE item_id LIKE '019%' OR item_id LIKE '025%'" 2>&1
echo ""

echo "[COMMAND 7]"
echo "backlogit shipment list"
echo "[OUTPUT]"
backlogit shipment list 2>&1
echo ""

echo "[COMMAND 8]"
echo "Get-ChildItem .backlogit/queue -Name"
echo "[OUTPUT]"
ls .backlogit/queue 2>&1
echo ""

echo "[COMMAND 9]"
echo "Get-ChildItem .backlogit/archive -Name | Select-String -Pattern '019|025|018'"
echo "[OUTPUT]"
ls .backlogit/archive 2>&1 | grep -E '019|025|018'
echo ""

echo "[COMMAND 10]"
echo "backlogit get 021-F"
echo "[OUTPUT]"
backlogit get 021-F 2>&1
echo ""

echo "[COMMAND 11]"
echo "backlogit get 021.001-T"
echo "[OUTPUT]"
backlogit get 021.001-T 2>&1
echo ""

echo "[COMMAND 12]"
echo "git status --porcelain"
echo "[OUTPUT]"
git status --porcelain 2>&1
echo ""

echo "[COMMAND 13]"
echo "git log --oneline -5"
echo "[OUTPUT]"
git log --oneline -5 2>&1
echo ""

echo "[COMMAND 14]"
echo "git symbolic-ref --short HEAD"
echo "[OUTPUT]"
git symbolic-ref --short HEAD 2>&1
echo ""

echo "[COMMAND 15]"
echo "Select-String -Path .backlogit/stash.jsonl -Pattern 'DB12DA37'"
echo "[OUTPUT]"
grep 'DB12DA37' .backlogit/stash.jsonl 2>&1
echo ""

echo "[COMMAND 16]"
echo "If file .backlogit/queue/025.001-T.md exists, run: Get-Content .backlogit/queue/025.001-T.md -Raw; otherwise search .backlogit for it"
echo "[OUTPUT]"
if [ -f ".backlogit/queue/025.001-T.md" ]; then
    cat ".backlogit/queue/025.001-T.md"
else
    find .backlogit -name "025.001-T.md" -exec cat {} \;
fi
echo ""

echo "[COMMAND 17]"
echo "Get-Content .backlogit/archive/018.008-T.md -Raw | head -n 30"
echo "[OUTPUT]"
if [ -f ".backlogit/archive/018.008-T.md" ]; then
    head -n 30 ".backlogit/archive/018.008-T.md"
else
    echo "File .backlogit/archive/018.008-T.md not found"
fi
echo ""

echo "[COMMAND 18]"
echo "backlogit --help"
echo "[OUTPUT]"
backlogit --help 2>&1
echo ""
