#!/bin/bash

# Konfiguration
API_URL="${API_URL:-http://localhost/api/shopping}"
AUTH_SERVICE="${AUTH_SERVICE:-http://localhost/api/auth}"
DEFAULT_USER_EMAIL="admin@thi.de"
DEFAULT_USER_PASSWORD="admin123"

# Step 1: Default User anlegen
echo "Creating default seed user..."
echo "Sending to $AUTH_SERVICE"
REGISTER_RESPONSE=$(curl -s -X POST "$AUTH_SERVICE/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$DEFAULT_USER_EMAIL\",\"password\":\"$DEFAULT_USER_PASSWORD\",\"role\":\"admin\"}")

echo "Register response: $REGISTER_RESPONSE"
echo ""

# Step 2: Login und Token holen
echo "Logging in as seed user..."
LOGIN_RESPONSE=$(curl -s -X POST "$AUTH_SERVICE/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$DEFAULT_USER_EMAIL\",\"password\":\"$DEFAULT_USER_PASSWORD\"}")

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "[ERROR] Failed to get JWT token"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

echo "[OK] Got JWT token"
echo ""

# Step 3: Produkt-Seeding mit Token
API_URL="$API_URL/products"
products=(
  '{"name":"Datenbrot","description":"Hausgemachtes Brot mit Analytics-Butter und frischen Kräutern","price":6.50,"category":"appetizer","image_url":"https://images.unsplash.com/photo-1509440159596-0249088772ff"}'
  '{"name":"Algorithmus-Antipasti","description":"Italienische Vorspeisen nach geheimem Algorithmus arrangiert","price":12.90,"category":"appetizer","image_url":"https://images.unsplash.com/photo-1544025162-d76694265947"}'
  '{"name":"Big Data Burger","description":"Saftig gegrilltes Rindfleisch mit Analytics-Sauce und Hashbrown-Pommes","price":16.90,"category":"main","image_url":"https://images.unsplash.com/photo-1568901346375-23c9450c58cd"}'
  '{"name":"Spaghetti Carbonara++","description":"Klassische Carbonara mit optimiertem Algorithmus für perfekten Geschmack","price":14.50,"category":"main","image_url":"https://images.unsplash.com/photo-1612874742237-6526221588e3"}'
  '{"name":"Cloud-Native Pizza","description":"Pizza Margherita mit containerisierten Zutaten und microservice-Gewürzen","price":13.90,"category":"main","image_url":"https://images.unsplash.com/photo-1513104890138-7c749659a591"}'
  '{"name":"Microservice Maultaschen","description":"Handgemachte Maultaschen mit verteilter Füllung nach Microservice-Prinzip","price":11.90,"category":"main","image_url":"https://images.unsplash.com/photo-1626200419199-391ae4be7a41"}'
  '{"name":"Container Curry","description":"Exotisches Curry in perfekt orchestrierten Container-Portionen","price":15.50,"category":"main","image_url":"https://images.unsplash.com/photo-1588166524941-3bf61a9c41db"}'
  '{"name":"Kubernetes Käsespätzle","description":"Schwäbische Spätzle mit automatisch skalierender Käsesauce","price":12.50,"category":"main","image_url":"https://images.unsplash.com/photo-1580013759032-c96505e24c1f"}'
  '{"name":"Tiramisu 2.0","description":"Traditionelles Tiramisu mit machine learning optimierter Mascarpone","price":7.50,"category":"dessert","image_url":"https://images.unsplash.com/photo-1571877227200-a0d98ea607e9"}'
  '{"name":"Docker Dampfnudeln","description":"Fluffige Dampfnudeln aus dem Container mit süßer Vanilla-Sauce","price":8.90,"category":"dessert","image_url":"https://images.unsplash.com/photo-1565958011703-44f9829ba187"}'
)

echo "Seeding products to $API_URL"
echo ""
for product in "${products[@]}"; do
  name=$(echo "$product" | jq -r '.name')
  echo "Creating: $name"
  response=$(curl -s -X POST "$API_URL" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "$product")
  if echo "$response" | jq -e '.id' > /dev/null 2>&1; then
    echo "[OK] Created successfully"
  else
    echo "[ERROR] Failed: $response"
  fi
  echo ""
done

echo "Seeding complete"