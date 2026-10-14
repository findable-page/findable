# findable logo drafts

Previews are contact sheets in `previews/`: `logos-light.png`, `logos-dark.png` (all three logos on #ffffff and #0f1115) and `marks-light.png`, `marks-dark.png` (each mark at 16, 32 and 180px). Dark previews swap the wordmark to its dark-mode color by hand, because the renderer does not apply prefers-color-scheme.

## A: Lens spark

A magnifier whose lens holds a four-point spark, so searching ends in an answer instead of a list of links. Colors: violet #6d5efc for the ring and handle, amber #ffb020 for the spark, wordmark #0f1115. Dark mode: logo files embed a style block that flips the wordmark to #f4f5f7 under prefers-color-scheme: dark; the violet and amber read on both backgrounds unchanged. Favicon: the ring and handle survive at 16px and the spark shrinks to a dot, which still reads as a magnifier. Strongest at small sizes of the three.

## B: Pin sparkle

A location pin with a round hole (the background shows through, so it works on any color) and an orange sparkle at the shoulder for the AI answer. Colors: teal #14b8a6 for the pin, orange #ff6a3d for the sparkle, wordmark #0f1115. Dark mode: same embedded style flip to #f4f5f7; teal and orange hold contrast on both. Favicon: the pin silhouette and hole read clearly at 16px, but the sparkle is only a speck there, so the mark is really just a pin (close to a maps icon).

## C: Beacon f

A rounded "f" whose hook ends in a beacon dot with one signal arc, meaning "we are broadcasting, find us". Colors: indigo violet #7c5cff for the f, orange #ff6a3d for the beacon, wordmark #0f1115. Dark mode: same embedded style flip to #f4f5f7. Favicon: the f is legible at 32px; at 16px the dot and arc blur into a smudge next to the f, so it is the weakest favicon. Most distinctive and ownable.

Notes: the wordmark is live text (Inter, then system-ui), not outlines. Viewbox width 270 leaves room for the wider fallback fonts, so with Inter there is a little slack on the right. Convert to paths before any final release.

## Pick

A: Lens spark. It is the only one that still reads as its own thing at 16px: the 32px and 16px cells in the marks previews keep ring, handle and a visible amber dot, while B collapses to a plain teal pin (indistinguishable from a maps app) and C smears the beacon into a smudge beside the f. The spark inside the lens says "search that ends in an answer", which is the product, and the warm amber on violet gives it the optimistic AI-era tone without copying the usual sparkle-on-a-chat-bubble. Violet and amber hold contrast on both #ffffff and #0f1115 with no color flip needed, so the mark is one asset for all contexts. The risk is that a magnifier is a generic search icon, but the spark is distinctive enough to carry it, and A is the strongest favicon of the three by a clear margin.

Tweak before release: at 16px the four-point spark loses its points and becomes a dot, so ship a dedicated favicon variant that fattens the spark (longer, blunter points, roughly 10 to 12 units across in the 64 viewbox instead of the current 22 wide thin star) and thins the ring stroke to 7 so the lens interior has room for it. Keep the full-size mark as drawn.
