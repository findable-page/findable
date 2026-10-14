import { ImageResponse, loadGoogleFont } from "workers-og";
import { scoreQuip } from "./quips";
import { type CardScores, shareCardHtml } from "./share-card";

export async function renderShareCard(scores: CardScores): Promise<Response> {
  const html = shareCardHtml(scores);
  const text = `findable${scores.host}0123456789/SEOAEG.page…${scoreQuip(scores.overall, scores.host)}`;
  const [regular, bold] = await Promise.all([
    loadGoogleFont({ family: "Inter", weight: 500, text }),
    loadGoogleFont({ family: "Inter", weight: 800, text }),
  ]);
  return new ImageResponse(html, {
    width: 1200,
    height: 630,
    fonts: [
      { name: "Inter", data: regular, weight: 500, style: "normal" },
      { name: "Inter", data: bold, weight: 800, style: "normal" },
    ],
  });
}
