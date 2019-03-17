package larsmonsen

import (
	"math/rand"
	"regexp"

	"github.com/go-chat-bot/bot"
)

const (
	pattern = "(?i)\\b(lars|monsen)\\b"
)

var (
	re          = regexp.MustCompile(pattern)
	monsenfacts = []string{
		"Lars Monsen er den eneste som har drept en grizzlybjørn med dens egne tenner.",
		"Lars Monsen har ikke med seg boksåpner på tur. Han spiser hermetikken hel.",
		"Ulvene gjør opp bål for å holde Monsen unna.",
		"Lars Monsen har hatt snøballkrig med seg selv i helvete.",
		"Lars Monsen gjør opp ild med bare hendene. Også i regnvær!",
		"Det eneste dyret Lars Monsen ikke kan lokke til seg med parringslyder er mammuter. Det er hovedsakelig fordi de er utryddet.",
		"Det amerikanske forsvaret brukte Lars Monsen som fasit da de utviklet GPS.",
		"Lars Monsen bygde varden på Kvigtind. Han kastet stein fra Tjokkelvatnet.",
		"Monsunperioden er oppkalt etter Lars Monsen.",
		"Lars Monsen er født med ski på beina, sekk på ryggen og en hundeslede med 12 hunder.",
		"Kvinner er fra Venus, menn fra Mars og Lars Monsen er fra Himalaya.",
		"Lars Monsen sover ikke, han venter… på at fisken skal nappe.",
		"Når Lars Monsen skal more seg, sitter han for seg selv midt på vidda, og forteller seg selv vitser han ikke har hørt før.",
		"Lars Monsen har klatret opp Trollveggen. Med én hånd.",
		"Lars har ugler i Monsen.",
		"Lars Monsen er 1.83 m høy. Når han reiser seg på bakbeina blir han 5.27 m høy.",
		"Lars Monsen har nok energi til å drive Las Vegas’ belysning i 2 år.",
		"Lars Monsen har ikke føtter. Han har fjellsko.",
		"Naturkreftene fikk kreftene av Lars Monsen.",
		"Det var Lars Monsen som gjorde opp ild på Sola.",
		"En gang spilte Lars Monsen golf. Da fikk han 18 hole-in-ones, og 8 ørreter på over 3 kilo i sidevannshinderet.",
		"En gang kom en bjørn inn i teltet til Lars Monsen og spiste maten hans. Da spiste Lars Monsen bjørnen.",
		"Da Challenger II-ekspedisjonen nådde bunnen av Marianegropen fant de Lars Monsen med våtdrakt, strikkharpun og syv inntil da ukjente arter i fangstnettet.",
		"Lars Monsen kan selge skinnet før bjørnen er skutt.",
		"Lars Monsen har ikke hunder for å trekke pulken. Han har dem med som nødproviant.",
		"Bjørnen går ikke i hi, den gjemmer seg for Lars Monsen.",
		"Kompasset viser ikke nord. Kompasset viser Lars Monsen.",
		"En gang, da Lars Monsen og Chuck Norris møttes i skogen, ropte Lars Monsen så høyt at Chuck Norris fikk rødt hår og fregner.",
		"Lars Monsen har det verken for varmt eller for kaldt. Han bryr seg ganske enkelt ikke om temperatur.",
		"Fjellvettreglene er en stiloppgave Lars Monsen skrev i første klasse på barneskolen.",
		"Lars Monsen danser med ulver.",
		"Lars Monsen er den eneste som alltid vinner stirrekonkurranser mot ugler.",
		"Ingen vet hvorfor Lars Monsen har kniv.",
		"Lars Monsen besto jegerprøven bare ved å møte opp.",
		"Ved oppdagelsen av Titanic-vraket fant de en stein på bunnen hvor det sto «Lars Monsen var her».",
		"Da Lars Monsen var marinejeger og dro på overlevelsesøvelse, kom han tilbake 7 kg tyngre enn når han dro.",
		"Supermann har en stor Lars Monsen-tatovering på brystkassa.",
		"Lars Monsen er så konge at både Sonja og Ari Behn kneler for han.",
		"Elver og innsjøer utvikler is på vinteren. Dette er for å beskytte fisken mot Lars Monsen.",
		"Ironisk nok tror ikke Lars Monsen på at det finnes utfordringer - det finnes bare pyser.",
		"Grunnen til at Lars Monsen har med seg gevær på jakt er for å kunne skyte varselskudd og gi bjørnen en sjanse. Når det ikke er bjørn i området jakter han heller ved å snike seg inn på dyrene bakfra og knekke nakken på de. Det gjelder også ryper og harer.",
		"Det er ikke tilfeldig at militæret bygger ned forsvaret i nord på samme tid som Lars Monsen har begynt å kjøre hundeløp der med jevne mellomrom.",
		"Hvis Lars Monsen har hendene fulle gjør han opp bål ved å kjefte på noen pinner til de tar fyr.",
		"Da Lars Monsen kappåt mot trollet, vant han overlegent uten sekk på magen. Deretter spiste han trollet til dessert.",
		"Lars Monsen er så rå at han lett kunne erstatte hele ferskvaredisken på Meny.",
		"Lars Monsen hugger ved med flat hånd.",
		"Det ryktes at Lars Monsen kom inn på 25. plass Finnmarksløpet, men han var faktisk på sin femte runde.",
		"Det finnes ikke ekko i fjellet, det er bare Lars Monsen som kødder med deg.",
		"Lars Monsen tegna kartet til Eirik Raude.",
		"Lars Monsen bruker sin egen pekefinger som jokkastikke.",
		"Alle kan kan pisse i motvind, men Lars Monsen blir ikke våt.",
		"IOC nekter Lars Monsen å stille i OL.",
		"Døden hadde en gang en nær Lars Monsen-opplevelse.",
		"Da Lars Monsen deltok i Verdensmesterskapet i poker vant hele greia med følgende hånd: Et visakort, et bibliotek-kort, en kvittering fra G-sport, et fiskekort fra Finnmarksvidda og et nøkkelkort tilhørende hotellrommet til en av damene som prøvde å sjekke han opp.",
		"Lars Monsen kan mette 5000 med en kniv, fem meter tau og en skog.",
		"I 2001 vant Lars Monsen World Rally Championship på New Zealand. Med en elg.",
		"En gang ble Lars Monsen bitt av en klapperslange. Etter fem timer med uutholdelige smerter døde slangen.",
		"Lars Monsen lærte maurene å bygge tua på sørsida av treet.",
		"Når en bjørn møter Lars Monsen begynner den å prate og håper Lars vil forsvinne av seg selv. Hvis ikke legger den seg med labbene over hodet, og spiller død.",
		"Da Lars Monsen feiret jul i Canada hadde han så lite mat at han koste seg med ferske rypespor til middag, is til dessert og en frostrøyk til kaffen.",
		"Eskimoer kaller eskimo-Rulle for Lars Monsen-Rulle.",
		"Om han hadde hatt interesse av det kunne Lars Monsen både fjerne himmelstrøk og fjerne østen.",
		"Lars Monsen veit at bringebær bare er blåbær i forhold til jordbær.",
		"Lars Monsen plantet Ibsens Ripsbusker og alle de andre buskvekstene.",
		"Lars Monsen får aldri kalde føtter.",
		"Lars Monsen står til Dovre faller.",
		"Uten mat og drikke, duger bare Lars Monsen.",
		"Lars Monsen svetter så mye på føttene at når han vrir om sokkene sine så går strømprisen ned.",
		"Så lenge det ikke blåser opp til snøstorm eller en million mygg forsøker å ete han synes Lars Monsen livet er enkelt. Resten av tida mener han er underholdende.",
		"Eskimoer har hundre ord for forskjellige typer snø. Lars Monsen har hundre ord for forskjellige typer eskimoer.",
		"Den einaste avgifta Lars Monsen betaler er bamsemoms.",
		"Godot venter på Lars Monsen.",
		"Har du prøvd å banke på baksiden av ytterdøra di? Hvis du gjør det, vil Lars Monsen komme å åpne. Utendørs er tross alt hans hjem.",
		"Monsen kan segla förutan vind. Monsen kan ro utan åror",
		"Lars Monsen lager pinneved av alle trær som våger seg over tregrensa.",
		"Nyere forskning viser at fiskelykke i liten grad handler om hell. Det avgjørende er hvor lenge det er siden Lars Monsen har vært der og fisket.",
		"En gang ble Lars Monsen angrepet av en krokodille, og laget en fin kajakk av den.",
		"I niendeklasse på ungdomsskolen kastet Lars Monsen en tennisball som fortsatt går i bane rundt jorden.",
		"Lars Monsen kan skru skruer i veggen med bare hendene.",
		"Da Gud sa: «La det bli lys», kunne man høre Lars Monsen svare «Det var virkelig på tide!».",
		"Lars Monsen kan nyse med øynene åpne.",
		"Historiene om Espen Askeladd er hovedsaklig basert på Lars Monsens selvopplevde eventyr.",
		"Lars Monsen synger kun i kano'n.",
		"En gang løp tiden fra Lars Monsen, men da stoppet den og ventet.",
		"Lars Monsens halvhøye La Sportiva lærstøvler er like steinharde som to femkilos ørreter i frysedisken.",
		"Lars Monsen synes voksenmenyen er barnemat.",
		"Lars Monsen er mektigere enn pennen.",
		"Lars Monsen har ingen problemer med å hoppe etter Wirkola.",
		"Lars Monsen var ikkje med i Noahs Ark. Han holdt pusten.",
		"Som deodorant bruker Lars Monsen Laphroaig Single Malt. Tripple Cask hvis det er fest.",
		"Lars Monsen er den eneste som kan gifte seg med Rein og komme unna med det.",
		"Ulvene forteller historier om valpen som ropte Lars Monsen.",
		"Det er bare én ting været ikke kan gjøre noe med, og det er Lars Monsen.",
		"Lars Monsen er den eneste som har banket opp noen med en kano.",
		"Lars Monsen går ikke å legger seg fordi det blir mørkt, det blir mørkt fordi Lars Monsen går og legger seg.",
		"Nylig oppdaget forskere ved MIT at Lars Monsen faktisk har 20x optisk zoom på øynene sine.",
		"Da Gud skapte himmelen og jorden ble Lars Monsen forbanna og lagde helvete.",
		"På en skala fra 1 til Lars Monsen er ikke Chuck Norris med en gang.",
		"Nye arkeologiske funn bekrefter teorien om at Lars Monses delte Rødehavet en gang rundt 1350 BC.",
		"Lars Monsen vet hvor haren hopper.",
		"Lars Monsen er sammen med Trine Rein. Tidligere har han bodd med Dinna Bjørn og Vigdis Hjort.",
		"En gang Lars Monsen var tom for proviant åt han en varde.",
		"Lars Monsen kan ta armhevinger med begge hendene på ryggen.",
		"Lars Monsen er den eneste som har fått både torsk og makrell i fjellvann.",
		"De eneste lovene Lars Monsen ikke respekterer er naturlovene.",
		"Det eneste som kan matche Northug på oppløpet er kameraet. Den eneste som kan slå de begge er Lars Monsen.",
		"Den eneste grunnen til at Lars Monsen har hodelykt, er at vi skal se hvor han er.",
		"Mayaindianerne forutså Lars Monsens utdrikkingslag 21. Desember 2012.",
	}
)

func larsmonsen(command *bot.PassiveCmd) (string, error) {
	if re.MatchString(command.Raw) {
		return monsenfacts[rand.Intn(len(monsenfacts))], nil
	}
	return "", nil
}

func init() {
	bot.RegisterPassiveCommand("larsmonsen", larsmonsen)
}
