#!/bin/bash

# Bulk Product Creator Script - Parallel version

API_URL="http://localhost:5007/api/v1/products"
HEALTH_URL="http://localhost:5007/health"
NUM_PRODUCTS=${1:-10}
PARALLEL_JOBS=${2:-10}

# Data arrays
CATEGORY_IDS=("507f1f77bcf86cd799439011" "507f1f77bcf86cd799439012" "507f1f77bcf86cd799439013" "507f1f77bcf86cd799439014" "507f1f77bcf86cd799439015" "507f1f77bcf86cd799439016" "507f1f77bcf86cd799439017" "507f1f77bcf86cd799439018" "507f1f77bcf86cd799439019" "507f1f77bcf86cd79943901a")
ADJECTIVES=("Amazing" "Premium" "Ultra" "Super" "Pro" "Elite" "Classic" "Modern" "Smart" "Eco" "Deluxe" "Compact" "Portable" "Wireless" "Digital" "Luxury" "Essential" "Advanced" "Professional" "Ultimate")
NOUNS=("Widget" "Gadget" "Device" "Tool" "Kit" "Set" "Pack" "Bundle" "System" "Machine" "Appliance" "Accessory" "Module" "Hub" "Controller" "Monitor" "Scanner" "Speaker" "Headphones" "Keyboard")
COLORS=("Red" "Blue" "Green" "Black" "White" "Silver" "Gold" "Purple" "Orange" "Pink" "Navy" "Teal" "Coral" "Crimson" "Indigo" "Violet" "Magenta" "Cyan" "Bronze" "Platinum")
BRANDS=("TechFlow" "NovaPro" "ZenithX" "ApexGear" "VeloCore" "PrimeTech" "NexGen" "QuantumX" "FusionMax" "EliteForce")

RESULTS_FILE=$(mktemp)
trap "rm -f $RESULTS_FILE" EXIT

create_one() {
    local idx=$1
    RANDOM=$((idx + $$))
    
    local cat_id=${CATEGORY_IDS[$RANDOM % ${#CATEGORY_IDS[@]}]}
    local adj=${ADJECTIVES[$RANDOM % ${#ADJECTIVES[@]}]}
    local noun=${NOUNS[$RANDOM % ${#NOUNS[@]}]}
    local color=${COLORS[$RANDOM % ${#COLORS[@]}]}
    local brand=${BRANDS[$RANDOM % ${#BRANDS[@]}]}
    
    local name="${brand} ${adj} ${color} ${noun}"
    local desc="Experience the ${adj} ${noun} by ${brand}. Premium ${color} finish."
    local price=$(( RANDOM % 990 + 10 )).$(( RANDOM % 100 ))
    local qty=$(( RANDOM % 100 + 1 ))
    local stock=$(( RANDOM % 490 + 10 ))
    local rating=$(( RANDOM % 10 + 1 ))
    
    local json="{\"categoryId\":\"${cat_id}\",\"name\":\"${name}\",\"description\":\"${desc}\",\"price\":${price},\"quantity\":${qty},\"stock\":${stock},\"rating\":${rating},\"imageUrl\":\"https://picsum.photos/400?r=${idx}\",\"photos\":[\"https://picsum.photos/400?r=${idx}a\"]}"
    
    local code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL" -H "Content-Type: application/json" -d "$json")
    
    if [ "$code" = "201" ] || [ "$code" = "200" ]; then
        echo "OK" >> "$RESULTS_FILE"
        echo "✅ #${idx}: ${name}"
    else
        echo "FAIL" >> "$RESULTS_FILE"
        echo "❌ #${idx}: Failed (HTTP ${code})"
    fi
}

echo "========================================"
echo "🚀 Bulk Product Creator"
echo "========================================"
echo "API: ${API_URL}"
echo "Products: ${NUM_PRODUCTS} | Workers: ${PARALLEL_JOBS}"
echo "========================================"

echo -n "Checking API... "
if curl -s -o /dev/null -w "%{http_code}" "$HEALTH_URL" | grep -q "200"; then
    echo "✅ OK"
else
    echo "⚠️ API may be down"
fi
echo ""

start_time=$(date +%s)

# Run in batches using background jobs
active_jobs=0
for i in $(seq 1 $NUM_PRODUCTS); do
    create_one $i &
    ((active_jobs++))
    
    if [ $active_jobs -ge $PARALLEL_JOBS ]; then
        wait -n 2>/dev/null || wait
        ((active_jobs--))
    fi
done
wait

end_time=$(date +%s)
duration=$((end_time - start_time))

success=$(grep -c "OK" "$RESULTS_FILE" 2>/dev/null || echo 0)
failed=$(grep -c "FAIL" "$RESULTS_FILE" 2>/dev/null || echo 0)

echo ""
echo "========================================"
echo "✅ Completed!"
echo "   Success: ${success} | Failed: ${failed}"
echo "   Time: ${duration}s"
if [ $duration -gt 0 ]; then
    rate=$(echo "scale=1; $NUM_PRODUCTS / $duration" | bc 2>/dev/null || echo "N/A")
    echo "   Rate: ${rate} products/sec"
fi
echo "========================================"
