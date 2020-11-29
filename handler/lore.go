package handler

import (
	"io"
	"net/http"
)

// Lore of Thor.
//
// This is just for fun, as we would like to have at least
// one tiny bit of functionality fitting to the project name.
//
// Source: https://norse-mythology.org/gods-and-creatures/the-aesir-gods-and-goddesses/thor/
const lore = `
Thor, the brawny thunder god, is the archetype of a loyal and honorable warrior, the ideal toward which the average human warrior aspired. He’s the indefatigable defender of the Aesir gods and their fortress, Asgard, from the encroachments of the giants, who are usually (although far from invariably) the enemies of the gods.

No one is better suited for this task than Thor. His courage and sense of duty are unshakeable, and his physical strength is virtually unmatched. He even owns an unnamed belt of strength (Old Norse megingjarðar) that makes his power doubly formidable when he wears the belt. His most famous possession, however, is his hammer, Mjöllnir (“Lightning”). Only rarely does he go anywhere without it. For the heathen Scandinavians, just as thunder was the embodiment of Thor, lightning was the embodiment of his hammer slaying giants as he rode across the sky in his goat-drawn chariot. (Of course, they didn’t believe he physically rode in a chariot drawn by goats – like everything else in Germanic mythology, this is a symbol used to express an invisible reality upon which the material world is perceived to be patterned.)

Thor’s particular enemy is Jormungand, the enormous sea serpent who encircles Midgard, the world of human civilization. In one myth, he tries to pull Jormungand out of the ocean while on a fishing trip, and is stopped only when his giant companion cuts the fishing line out of fear. Thor and Jormungand finally face each other during Ragnarok, however, when the two put an end to each other.

Given his ever-vigilant protection of the ordered cosmos of pre-Christian northern Europe against the forces of chaos, destruction, and entropy represented by the giants, it’s somewhat ironic that Thor is himself three-quarters giant. His father, Odin, is half-giant, and his mother, variously named as Jord (Old Norse “Earth”), Hlöðyn, or Fjörgyn, is entirely of giant ancestry. However, such a lineage is very common amongst the gods, and shows how the relationship between the gods and the giants, as tense and full of strife as it is, can’t be reduced to just enmity.

His activities on the divine plane were mirrored by his activities on the human plane (Midgard), where he was appealed to by those in need of protection, comfort, and the blessing and hallowing of places, things, and events. Numerous surviving runic inscriptions invoke him to hallow the words and their intended purpose, and it was he who was called upon to hallow weddings.[4] (Evidence of this is preserved, amongst other places, in the tale of Thor Disguised as a Bride.) The earliest Icelandic settlers implored him to hallow their plot of land before they built buildings or planted crops.[5]

Thor’s hammer could be used to hallow as readily as it could be used to destroy – and, in effect, these two properties were one and the same, since any purification necessarily involves the banishing of hostile forces or elements. The blessing of weddings, for example, was effected through his hammer. Perhaps the most striking case of this, however, is his ability to kill and eat the goats that drive his chariot, gather their bones together in their hides, bless the hides with the hammer, and bring the animals back to life, as healthy and vital as before.[6]

In addition to his role as a model warrior and defender of the order of society and its ambitions, Thor also played a large role in the promotion of agriculture and fertility (something which has already been suggested by his blessing of the lands in which the first Icelanders settled). This was another extension of his role as a sky god, and one particularly associated with the rain that enables crops to grow. As the eleventh-century German historian Adam of Bremen notes, “Thor, they say, presides over the air, which governs the thunder and lightning, the winds and rains, fair weather and crops.”[7] His seldom-mentioned wife, Sif, is noted for her golden hair above all else, which is surely a symbol for fields of grain. Their marriage is therefore an instance of what historians of religion call a “hierogamy” (divine marriage), which, particularly among Indo-European peoples, generally takes place between a sky god and an earth goddess. The fruitfulness of the land and the concomitant prosperity of the people is a result of the sexual union of sky and earth.[8]

Through archaeological evidence, the veneration of Thor can be traced back as far as the Bronze Age,[9] and his cult has gone through numerous permutations across time and space. One of the features that remained constant from the Bronze Age up through the Viking Age, however, is Thor’s role as the principal deity of the second class or “function” of the three-tiered social hierarchy of traditional European society – the function of warriors and military strength. (The first function was that of rulers and sovereignty, and the third was that of farmers and fecundity.)[10]

Thor seems to have always had close ties to the third function as well as the second, and during the Viking Age, a time of great social confusion and innovation, this connection with the third function seems to have been strengthened still more. This made him the foremost god of the common people in Scandinavia and the viking colonies.[11]

This role can be made clearer by contrasting Thor with the god who was virtually his functional opposite: Odin. Odin was the foremost deity appealed to by rulers, outcasts, and “elite” persons of every sort. Odin’s primary values are quite rarefied: ecstasy, knowledge, magical power, and creative agency. They stand in stark contrast to Thor’s more homely virtues. The Eddas and sagas portray the relationship between the two gods as being often uneasy as a result. At one point, Odin taunts Thor: “Odin’s are the nobles who fall in battle, but Thor’s are the thralls.”[12] In another episode, Odin is conferring blessings upon a favored hero of his, Starkaðr, and each blessing is matched by a curse from Thor. In the most telling example, Odin grants Starkaðr the favor of the nobility and rulers, while Thor declares that he will always be scorned by the commoners.[13]

Due to demographic shifts, whereby the second and third functions became largely indistinguishable from one another, the prominence of Thor seems to have increased at the expense of Odin throughout the Viking Age (c. 793-1000 CE). Late period sources describe Thor as the foremost of all the Aesir,[14] a statement that would have been rather ludicrous before the Viking Age, when Odin and his Anglo-Saxon and continental equivalents occupied this position.

Nowhere was this trend more pronounced than in Iceland, which was originally settled in the ninth century by farming colonists fleeing what they found to be the oppressive and arbitrary rule of an Odin-worshiping Norwegian king. The sagas are rife with examples of the fervent veneration of Thor amongst the Icelanders, and in the Landnámabók, the Icelandic “Book of Settlements,” roughly a quarter of the four thousand people mentioned in the narrative have Thor’s name or a clear allusion to him somewhere in their own names. Famed Old Norse scholar E.O.G. Turville-Petre admirably summarizes: “In these [late Viking Age Icelandic] sources Thor appears not only as the chief god of the settlers but also as patron and guardian of the settlement itself, of its stability and law.”

There’s yet another reason for the upsurge in the worship of Thor during the Viking Age. When Christianity first reached Scandinavia and the viking colonies, the people tolerated the cult of the new god just like they tolerated the cult of any other god. However, when it became clear that the Christians had no intention of extending this same tolerance to those who continued to adhere to the worship of the old gods, but instead wanted to eradicate the traditional religion of northern Europe and its accompanying way of life and replace it with a foreign religion, the northern Europeans retaliated. And who better to defend their traditional way of life and worldview from hostile, invading forces than Thor? One of the many areas of life in which this struggle manifested – and one of the easiest to trace by the methods of modern anthropology – was modes of dress. In deliberate contrast to the cross amulets that the Christians wore around their necks, those who continued to follow the old ways started to wear miniature Thor’s hammers around their necks. Archaeological discoveries of these hammer pendants are concentrated in precisely the areas where Christian influence was the most pronounced. Though ultimately doomed, their efforts to preserve their ancestral traditions no doubt benefited from the divine patron whom they could look to as a model.
`

func Lore() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, lore)
	}
}
