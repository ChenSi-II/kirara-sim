package keys

import (
	"encoding/json"
	"errors"
	"strings"
)

type Set int

func (s *Set) MarshalJSON() ([]byte, error) {
	return json.Marshal(setNames[*s])
}

func (s *Set) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	str = strings.ToLower(str)
	for i, v := range setNames {
		if v == str {
			*s = Set(i)
			return nil
		}
	}
	return errors.New("unrecognized set key")
}

func (s Set) String() string {
	return setNames[s]
}

var setNames = []string{
	"",
	"adaycarvedfromrisingwinds",
	"adventurer",
	"artifact15048",
	"archaicpetra",
	"aubadeofmorningstarandmoon",
	"berserker",
	"blizzardstrayer",
	"bloodredproof",
	"bloodstainedchivalry",
	"braveheart",
	"celestialgift",
	"crimsonwitchofflames",
	"deepwoodmemories",
	"defenderswill",
	"desertpavilionchronicle",
	"disenchantmentindeepshadow",
	"echoesofanoffering",
	"emblemofseveredfate",
	"finaleofthedeepgalleries",
	"flowerofparadiselost",
	"fragmentofharmonicwhimsy",
	"gambler",
	"glacierandsnowfield",
	"gladiatorsfinale",
	"gildeddreams",
	"goldentroupe",
	"heartofdepth",
	"huskofopulentdreams",
	"instructor",
	"lavawalker",
	"longnightsoath",
	"luckydog",
	"maidenbeloved",
	"marechausseehunter",
	"martialartist",
	"nightoftheskysunveiling",
	"nighttimewhispersintheechoingwoods",
	"noblesseoblige",
	"nymphsdream",
	"obsidiancodex",
	"oceanhuedclam",
	"paleflame",
	"prayersfordestiny",
	"prayersforillumination",
	"prayersforwisdom",
	"prayerstospringtime",
	"resolutionofsojourner",
	"retracingbolide",
	"scholar",
	"scrolloftheheroofcindercity",
	"shimenawasreminiscence",
	"silkenmoonsserenade",
	"songofdayspast",
	"tenacityofthemillelith",
	"theexile",
	"thunderingfury",
	"thundersoother",
	"tinymiracle",
	"travelingdoctor",
	"unfinishedreverie",
	"vermillionhereafter",
	"viridescentvenerer",
	"vourukashasglow",
	"wandererstroupe",
}

const (
	NoSet Set = iota
	ADayCarvedFromRisingWinds
	Adventurer
	Artifact15048
	ArchaicPetra
	AubadeOfMorningstarAndMoon
	Berserker
	BlizzardStrayer
	BloodRedProof
	BloodstainedChivalry
	BraveHeart
	CelestialGift
	CrimsonWitchOfFlames
	DeepwoodMemories
	DefendersWill
	DesertPavilionChronicle
	DisenchantmentInDeepShadow
	EchoesOfAnOffering
	EmblemOfSeveredFate
	FinaleOfTheDeepGalleries
	FlowerOfParadiseLost
	FragmentOfHarmonicWhimsy
	Gambler
	GlacierAndSnowfield
	GladiatorsFinale
	GildedDreams
	GoldenTroupe
	HeartOfDepth
	HuskOfOpulentDreams
	Instructor
	Lavawalker
	LongNightsOath
	LuckyDog
	MaidenBeloved
	MarechausseeHunter
	MartialArtist
	NightOfTheSkysUnveiling
	NighttimeWhispersInTheEchoingWoods
	NoblesseOblige
	NymphsDream
	ObsidianCodex
	OceanHuedClam
	PaleFlame
	PrayersForDestiny
	PrayersForIllumination
	PrayersForWisdom
	PrayersToSpringtime
	ResolutionOfSojourner
	RetracingBolide
	Scholar
	ScrollOfTheHeroOfCinderCity
	ShimenawasReminiscence
	SilkenMoonsSerenade
	SongOfDaysPast
	TenacityOfTheMillelith
	TheExile
	ThunderingFury
	Thundersoother
	TinyMiracle
	TravelingDoctor
	UnfinishedReverie
	VermillionHereafter
	ViridescentVenerer
	VourukashasGlow
	WanderersTroupe
)
