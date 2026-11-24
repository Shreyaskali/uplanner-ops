// parity-js emits the same normalised fit result as parity-go, using the
// web tool's compass-core.js. Run with the University Planner as cwd:
//
//	node /path/to/uplanner-ops/parity/js-engine.mjs fixture.json
import { createRequire } from 'node:module';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const require = createRequire(import.meta.url);
const corePath = process.env.COMPASS_CORE || resolve(process.cwd(), 'compass-core.js');
const core = require(corePath);

const fixturePath = process.argv[2];
if (!fixturePath) {
  console.error('usage: node js-engine.mjs <fixture.json>');
  process.exit(2);
}
const fx = JSON.parse(readFileSync(fixturePath, 'utf8'));
const p = core.buildProfile(fx.profile || {});
const a = core.analyse(fx.jd, null, { profile: p });

const result = {
  name: fx.name,
  ok: a.ok,
  score: a.score,
  band: a.band,
  confidence: a.confidence,
  dimensions: a.dimensions.map((d) => ({
    id: d.id, label: d.label, group: d.group, fit: d.fit,
    demand: d.demand, evidence: d.evidence, roles: d.roles, terms: d.terms,
  })),
  requirements: a.requirements,
  matchedKeywords: a.matchedKeywords,
};
console.log(JSON.stringify(result));
