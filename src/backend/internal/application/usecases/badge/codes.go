package badge

// Коды бейджей — совпадают с Badge.code в справочнике БД.
// Хранятся как константы, чтобы не путаться в строках по коду.
const (
	// Личные
	CodeFirstStone    = "badge_first_stone"    // первая партия
	CodeFirstBlood    = "badge_first_blood"    // первая победа
	CodeLizard        = "badge_lizard"         // 3 победы
	CodeMarathoner    = "badge_marathoner"     // 10 партий
	CodeStoneEater    = "badge_stone_eater"    // 10 побед
	CodeFire          = "badge_fire"           // стрик 3
	CodeUnstoppable   = "badge_unstoppable"    // стрик 5
	CodeSixSeven      = "badge_six_seven"      // стрик 6/7
	CodeHeadhunter    = "badge_headhunter"     // снял чужой стрик
	CodeLucky         = "badge_lucky"          // победа с разницей 0.5
	CodeFriendlyMatch = "badge_friendly_match" // сыграл с другим клубом
	CodeGlobalist     = "badge_globalist"      // сыграл с другим городом

	// Сезонные личные
	CodeSeasonClubGold   = "badge_season_club_gold"
	CodeSeasonClubSilver = "badge_season_club_silver"
	CodeSeasonClubBronze = "badge_season_club_bronze"
	CodeSeasonDivGold    = "badge_season_div_gold"
	CodeSeasonDivSilver  = "badge_season_div_silver"
	CodeSeasonDivBronze  = "badge_season_div_bronze"

	// Супертитул
	CodeSuperChampion67 = "badge_super_champion_67"
)
