package immutable

import (
  "bytes"
  "fmt"
  "strconv"
  "strings"
)

type myValue struct {
  key   uint
  value string
}

func (e *myValue) Hash() uint { return e.key }
func (e *myValue) Equal(b Value) bool {
  if b, ok := b.(*myValue); ok {
    return e.key == b.key
  }
  return false // for comparison with myCollidingValue
}
func (e *myValue) String() string {
  return fmt.Sprintf("{%d => %#v}", e.key, e.value)
}

type myCollidingValue struct {
  key   uint
  value string
}

func (e *myCollidingValue) Hash() uint { return e.key }
func (e *myCollidingValue) Equal(b Value) bool {
  // Always false unless the exact same value is provided.
  // This makes testing collision easy.
  return e == b
}
func (e *myCollidingValue) String() string {
  return fmt.Sprintf("{%d => %#v}", e.key, e.value)
}

// newValue creates a new myValue using path for both the key and value.
func newValue(path string) *myValue {
  return &myValue{buildHamtKey(path), path}
}

func newCollidingValue(keypath, value string) *myCollidingValue {
  return &myCollidingValue{buildHamtKey(keypath), value}
}

// buildHamtKey takes a string of slash-separated integers and builds a key
// where each integer maps to one level of branching in CHAMP.
//
// For instance, the key "1/2/3/4" produces the key:
//   0b000100_000011_000010_000001
//          4      3      2      1
//
func buildHamtKey(path string) uint {
  paths := strings.Split(path, "/")
  key := uint(0)
  shift := uint(0)
  for _, p := range paths {
    index, _ := strconv.Atoi(p)
    key |= uint(index) << shift
    shift += hamtBits
  }
  return key
}

func hashFNV1a(s string) uint32 {
  const prime uint32 = 0x01000193 // pow(2,24) + pow(2,8) + 0x93
  hash := uint32(0x811C9DC5)      // seed
  for i := 0; i < len(s); i++ {
    hash = (uint32(s[i]) ^ hash) * prime
  }
  return hash
}

func hashFNV1aUint32(v uint32) uint32 {
  const prime uint32 = 0x01000193 // pow(2,24) + pow(2,8) + 0x93
  hash := uint32(0x811C9DC5)      // seed
  hash = (v&0x000000ff ^ hash) * prime
  hash = (v&0x0000ff00 ^ hash) * prime
  hash = (v&0x00ff0000 ^ hash) * prime
  return (v&0xff000000 ^ hash) * prime
}

func hashFNV1aUint64(v uint64) uint64 {
  const prime uint64 = 0x100000001B3 // pow(2,40) + pow(2,8) + 0xb3
  hash := uint64(0xCBF29CE484222325) // seed
  hash = (v&0x00000000000000ff ^ hash) * prime
  hash = (v&0x000000000000ff00 ^ hash) * prime
  hash = (v&0x0000000000ff0000 ^ hash) * prime
  hash = (v&0x00000000ff000000 ^ hash) * prime
  hash = (v&0x000000ff00000000 ^ hash) * prime
  hash = (v&0x0000ff0000000000 ^ hash) * prime
  hash = (v&0x00ff000000000000 ^ hash) * prime
  return (v&0xff00000000000000 ^ hash) * prime
}

func hashFNV1aUint(v uint) uint {
  if intSize == 64 {
    return uint(hashFNV1aUint64(uint64(v)))
  }
  return uint(hashFNV1aUint32(uint32(v)))
}

// FmtKey formats a key into a slash-separated path of subkeys.
//
// E.g. key 0b1011_001010_001001_001000_000111_000110_000101_000100_000011_000010_000001
// returns "1/2/3/4/5/6/7/8/9/10/11"
//
// E.g. key 0b000101_000000_000000_000010_000001
// returns "1/2/0/0/5" (note lack of trailing zeroes)
//
func FmtKey(key uint) string {
  v := make([]string, 0, 10)
  lastNonZeroIndex := 0
  i := 0
  for shift := uint(0); shift < hamtBranches; shift += hamtBits {
    index := int((key >> shift) & hamtMask)
    if index != 0 {
      lastNonZeroIndex = i
    }
    v = append(v, strconv.Itoa(index))
    i++
  }
  return strings.Join(v[:lastNonZeroIndex+1], "/")
}

// fmtbmap formats a hamt bitmap as a string, grouped by hamtBits number of bits.
// When hamtBranches is 32, the output looks like this:
//   "01_00001_00001_00001_00001_00001_00001"
// When hamtBranches is 64, the output looks like this:
//   "0001_000001_000001_000001_000001_000001_000001_000001_000001_000001_000001"
//
func fmtbmap(u uint) string {
  var buf bytes.Buffer
  if hamtBranches == 64 {
    buf.Grow(128) // with extra '_' at every hamtBits byte
    fmt.Fprintf(&buf, "%064b", u)
  } else if intSize == 64 && u > 0xFFFFFFFF {
    return fmtbits(u)
  } else {
    buf.Grow(64) // with extra '_' at every hamtBits byte
    fmt.Fprintf(&buf, "%032b", u)
  }
  b := buf.Bytes()
  srci := len(b) - 1
  b = b[:cap(b)]
  dsti := len(b) - 1
  nbit := uint(0)
  for srci >= 0 {
    b[dsti] = b[srci]
    dsti--
    nbit++
    if nbit == hamtBits {
      b[dsti] = '_'
      dsti--
      nbit = 0
    }
    srci--
  }
  return string(b[dsti+1:])
}

func fmtbits(u uint) string {
  var s string
  if intSize == 64 {
    s = fmt.Sprintf("%064b", u)
  } else {
    s = fmt.Sprintf("%032b", u)
  }
  var s2 string
  for i := 0; i < len(s); i += 8 {
    if i > 0 {
      s2 += " "
    }
    s2 += s[i : i+8]
  }
  return s2
}

// random entries from test_colors.txt
var testDataColorNames = []string{
  "Monks Robe",
  "Hint of Spring Burst",
  "Spring Lily",
  "Turned Leaf",
  "Wavelet",
  "Rose Water",
  "Satin Green",
  "Midnight Violet",
  "Light Salome",
  "Tan Whirl",
  "Acapulco Cliffs",
  "Matt Sage",
  "Secrecy",
  "Lemon Chiffon",
  "Windjammer",
  "Violet Persuasion",
  "Conch Shell",
  "Palm Leaf",
  "Burled Redwood",
  "Pirate Gold",
  "Foggy Quartz",
  "Blackheath",
  "Silkie Chicken",
  "Burnt Sienna",
  "Delta Waters",
  "Dark Gold",
  "Beanstalk",
  "Pink Dazzle",
  "Peeled Asparagus",
  "Hint of Daly Waters",
  "Mauve Wisp",
  "Oriental Spice",
  "Florentine Lapis",
  "Fondue Fudge",
  "Mission Jewel",
  "Mauve Stone",
  "Rose",
  "Diva Blue",
  "Oakmoss",
  "Calamansi Green",
  "Blue Bay",
  "Sea Sprite",
  "Softsun",
  "Bluish Green",
  "Tropical Teal",
  "Eggshell Paper",
  "Peruvian Soil",
  "Aloe Essence",
  "Light Medlar",
  "Feta",
  "Curious Blue",
  "Faded Rose",
  "Tibetan Temple",
  "Statue of Liberty",
  "Young At Heart",
  "Gaia",
  "Viking Castle",
  "Elf Slippers",
  "Monastir",
  "Tetsu Green",
  "Yolk Yellow",
  "Barbecue",
  "Elmer's Echo",
  "Bitter Chocolate",
  "Bright Blue",
  "Depth Charge",
  "Waywatcher Green",
  "Bypass",
  "Bleu Nattier",
  "Overcast Sky",
  "Pumpkin Seed",
  "Spring Bouquet",
  "Pinball",
  "Delicate White",
  "Alaskan Blue",
  "Vulcan Burgundy",
  "Virtuous",
  "Cane Sugar",
  "Stoic White",
  "Cab Sav",
  "Amberlight",
  "Light Blue Stream",
  "Persimmon Orange",
  "Dark Royal Blue",
  "Catmint",
  "Blue et une Nuit",
  "Sponge Cake",
  "Blue Expanse",
  "Sunlight",
  "Pink Pleasure",
  "Sea Salt",
  "Stonegate",
  "Princess Pink",
  "Potting Soil",
  "Canyon Cloud",
  "Old Truck",
  "Peach Flower",
  "Sacred Turquoise",
  "Bio Blue",
  "Moorland Heather",
  "Top Hat Tan",
  "Freefall",
  "Peach Cobbler",
  "Emperor Jade",
  "Electric Green",
  "Castle Stone",
  "Double Duty",
  "Cheddar Biscuit",
  "Urobilin",
  "Chinchilla",
  "Alverda",
  "Dust",
  "Riviera Sea",
  "Sandy Tan",
  "Golden Cream",
  "Moby Dick",
  "Hawaiian Sunset",
  "Wisteris",
  "Pale",
  "Tiara",
  "Northern Lights",
  "Glade",
  "Blue Sail",
  "Coral Dusk",
  "Stone Harbour",
  "Citadel",
  "Pea Green",
  "Medium Spring Green",
  "Taupe Grey",
  "Vino Tinto",
  "Rainy Day",
  "Earthbound",
  "Bright Magenta",
  "Ripe Rhubarb",
  "Hint of Ghost Town",
  "Wicker Basket",
  "Melanzane",
  "Chloride",
  "Heavenly",
  "Barbados",
  "White Heat",
  "Blue Plate",
  "Tranquil Bay",
  "Mustard Yellow",
  "Gentian Violet",
  "Plain and Simple",
  "Siesta Rose",
  "Capital Blue",
  "Kobe",
  "Turkish Rose",
  "Governor Bay",
  "Dark Shamrock",
  "Waterfall",
  "Koopa Green Shell",
  "Old Bear",
  "Tropical Violet",
  "Wild Ginseng",
  "Méi Gūi Hóng Red",
  "Forbidden Fruit",
  "Hint of Soya",
  "Averland Sunset",
  "Medium Orchid",
  "Ricochet",
  "Orange Roughy",
  "Yawl",
  "Battle Dress",
  "Crusta",
  "Light Green Glacier",
  "Dull Gold",
  "Flagstone",
  "Coronado Moss",
  "Shower",
  "Grand Purple",
  "Phenomenon",
  "Frozen Moss Green",
  "Clay Fire",
  "Butterscotch Ripple",
  "Forest Blues",
  "Jetski Race",
  "Dune Shadow",
  "Redtail",
  "Honesty",
  "Whiskey",
  "Cor-de-pele",
  "Teal Tune",
  "Birōdo Green",
  "Mint Chiffon",
  "Genetic Code",
  "Polished Copper",
  "Floral Leaf",
  "Private Tone",
  "Bali Batik",
  "Black Onyx",
  "Harbour Afternoon",
  "Hint of Starlight Blue",
  "Buff Orange",
  "Flashman",
  "Lone Pine",
  "Spray of Mint",
  "400XT Film",
  "Tangerine Skin",
  "Mouse Catcher",
  "Cowgirl Boots",
  "Spinach Soup",
  "Light Blue Cloud",
  "Naval",
  "Roebuck",
  "Flesh",
  "Lounge Leather",
  "Light Zenith Heights",
  "Pigeon",
  "Hidden Waters",
  "Blushing Bride",
  "Cobblestone Street",
  "City Tower",
  "Anonymous",
  "Glacier Green",
  "Stil De Grain Yellow",
  "Scarlet",
  "Pearl Ash",
  "April Wedding",
  "Steel Teal",
  "Mischka",
  "Pistachio Tang",
  "Craft",
  "Smoky Black",
  "Dead Forest",
  "Misty Violet",
  "Yuè Guāng Lán Blue",
  "Lemon Caipirinha",
  "Primo",
  "Victorian Crown",
  "Dusty Coral",
  "Dull Purple",
  "Blue By You",
  "Simply Elegant",
  "Hint of Tenzing",
  "Golden Harmony",
  "Synthetic Spearmint",
  "Pontoon",
  "Cool Slate",
  "Apricot Mousse",
  "Amethyst Show",
  "Les Cavaliers Beach",
  "Sanskrit",
  "Abstract White",
  "Modern Monument",
  "Carrot",
  "Vibrant Green",
  "Photo Grey",
  "Sail On",
  "Guiding Star",
  "Nymph's Delight",
  "Chinois Green",
  "Violet Ice",
  "Blue Hour",
  "Oregon Hazel",
  "Woolly Beige",
  "Aloof",
  "Spiced Nectarine",
  "Antique Wicker Basket",
  "Blue Graphite",
  "Cosmo Purple",
  "Anchor Point",
  "Blue Martina",
  "Italian Lace",
  "Slices of Happy",
  "Nature's Delight",
  "Island Oasis",
  "Plum Purple",
  "Grey Monument",
  "Mulberry Bush",
  "Flour Sack",
  "Frosted Sugar",
  "Cendre Blue",
  "Mazzone",
  "Trite White",
  "Barren",
  "Saddle Brown",
  "Vegetation",
  "Sesame",
  "Hint of Green Frost",
  "Night Rendezvous",
  "Young Night",
  "Aqua Smoke",
  "Barrel Stove",
  "Chalk Pink",
  "Happy Daze",
  "Mauve Chalk",
  "Underhive Ash",
  "Sizzling Sunrise",
  "Guppy Violet",
  "Toreador",
  "Grey Suit",
  "Pink Gin",
  "Ottoman",
  "Illuminating Emerald",
  "Lemon Chiffon Pie",
  "Organic Bamboo",
  "Morocco",
  "Elegant Ivory",
  "Mango Salsa",
  "Mocha Mousse",
  "Paris Daisy",
  "White Whale",
  "Cinnamon Brandy",
  "Whispering Winds",
  "Glazed Pears",
  "Deep Storm",
  "Petite Purple",
  "Henna",
  "Aqua Tint",
  "Vivid Lime Green",
  "Light Elusive Mauve",
  "Tea Leaf Brown",
  "Blue Oasis",
  "Tree Peony",
  "Yellow Yarn",
  "Walleye",
  "Bengara Red",
  "Light Washed Blue",
  "Belle of the Ball",
  "Teeny Bikini",
  "First Frost",
  "Aqua Vitale",
  "Brick Orange",
  "Smoked Purple",
  "Beau Blue",
  "Light Carmine Pink",
  "Poplar",
  "Leticiaz",
  "Salmon Smoke",
  "Estate Blue",
  "Light Brown Sugar",
  "Calliste Green",
  "Hint of Green Wash",
  "British Shorthair",
  "Parakeet Pete",
  "La Rioja",
  "Curd",
  "Coral Cloud",
  "Cloud Dancer",
  "Hint of Pink Polar",
  "Rustic Brown",
  "Persian Blue",
  "Dutch Jug",
  "Salt Water Taffy",
  "Valencia",
  "Red Mahogany",
  "Plumburn",
  "Cinnamon Sand",
  "Bladed Grass",
  "Sultry Sea",
  "Tort",
  "Light Carolina",
  "Ground Cover",
  "Deep Mauve",
  "Ebb",
  "Blue Planet",
  "Lit'L Buoy Blew",
  "Bonjour",
  "Emberglow",
  "Lavender Ash",
  "Narvik",
  "Light Favourite Lady",
  "Afterglow",
  "Golgfag Brown",
  "Upbeat",
  "Huáng Sè Yellow",
  "California Wine",
  "Angel Blue",
  "Lavender Mist",
  "Lavender Indigo",
  "Snip of Tannin",
  "Moss Island",
  "Coastal Fjord",
  "Winter Nap",
  "Matisse",
  "Seneca Rock",
  "Garden Party",
  "Blue Prince",
  "Dizzy Days",
  "Orchilla",
  "Light Katsura",
  "Stanley",
  "Wake Me Up",
  "Deserted Path",
  "Lemon Slice",
  "Slipper Satin",
  "Dover Grey",
  "Limerick",
  "Bile",
  "Lemonwood Place",
  "Special Delivery",
  "Miami Jade",
  "Tropical Wood",
  "Passion Potion",
  "Sweet Tooth",
  "Dark Border",
  "Tombstone Grey",
  "Olive Yellow",
  "Russian Green",
  "Kangaroo Pouch",
  "Barley White",
  "Whisper",
  "Cerise Red",
  "Golden Granola",
  "Gratefully Grass",
  "Blue Smart",
  "Allure",
  "Monarch",
  "Physalis Peal",
  "Apple Green",
  "Safflower Bark",
  "Popcorn",
  "Plum Wine",
  "Spice Is Nice",
  "Ketchup",
  "Just A Tease",
  "Light Shetland Lace",
  "Vermilion",
  "Toscana",
  "Lye",
  "Sweet Grape",
  "Bubbles",
  "Lake Blue",
  "Deep Taupe",
  "Calla Green",
  "Cream Blush",
  "Cocoa Bean",
  "Passionate Pause",
  "Alizarin",
  "Nice Blue",
  "Penelope Pink",
  "Lunar Eclipse",
  "Unmellow Yellow",
  "More Maple",
  "C-3PO",
  "Orange Delight",
  "New Clay",
  "Potter's Pink",
  "Lucky Point",
  "Pure Zeal",
  "Cryptic Light",
  "Reeds",
  "Love Spell",
  "North Rim",
  "Pink Peacock",
  "Soft Blue",
  "Water Droplet",
  "Lavender Earl",
  "Buttered Popcorn",
  "Uncharted",
  "Midnight Moss",
  "Red Berry",
  "Light Periwinkle",
  "Rainbow Bright",
  "Cameo Role",
  "Tetsu-Kon Blue",
  "Blue Shadow",
  "Embroidered Silk",
  "Hemp Tea",
  "Holy Fern",
  "Electric Ultramarine",
  "Vers de Terre",
  "Spring Crocus",
  "Autumn Bark",
  "Burning Sand",
  "Tattered Teddy",
  "Namakabe Brown",
  "Sonora Rose",
  "Liberty",
  "Mirabella",
  "Ooid Sand",
  "Caribou Herd",
  "Secrent of Mana",
  "Tabasco",
  "Ecstasy",
  "Coral Coast",
  "Ocean Cruise",
  "Rancho Verde",
  "Winter Willow Green",
  "Atlantic Mystique",
  "Drops of Honey",
  "Light Timeless",
  "Shinbashi",
  "At The Beach",
  "Bunny Hop",
  "Hint of Raw Cotton",
  "Gettysburg Grey",
  "Solar Power",
  "Observatory",
  "Greener Grass",
  "Fossil Stone",
  "Molly Green",
  "Clay Bath",
  "Wageningen Green",
  "Male Betta",
  "Oyster Pink",
  "Funk",
}
