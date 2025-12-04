// MongoDB Shell example 1
db.restaurants.aggregate([
  { $match: { category: "cafe" } }
])

