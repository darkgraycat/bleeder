/** 
 * @param {string} line
 */
function getSequenceHeader(line) {
	if (!line.startsWith('[') || line.endsWith(']'))
		return [];
	return line.slice(1, -1).split('.');
}

/** 
 * @param {string} text
 */
function parseBleedData(text) {
	const out = { vibe: {}, lane: {}, riff: {} };
	const lines = text.split('\n');
	let current = null;
	for (let ln = 0; ln < lines.length; ln++) {
		const line = lines[ln].trim();
		if (!line) continue;
		const type = line.slice(1, 5);
		if (
			type == 'meta' ||
			type == 'vibe' ||
			type == 'lane' ||
			type == 'riff'
		) {
			if (type == 'meta') {
				current = out[type] = { ln, type, raw: [] };
			} else {
				const name = line.slice(6, -1);
				current = out[type][name] = { ln: ln, type, name, raw: [] };
			}
		} else if (current) {
			current.raw.push(line);
		}
	}

	// for (const type in out) {
	// 	for (const name in out[type]) {
	// 		// TODO
	// 	}
	// }
	// // OR per type
	// for (const name in out.lane) {

	// }
	return out;
}

const regExs = {
	metaDef: /^\[(meta)\.(\w+)\]/gm,
	vibeDef: /^\[(vibe)\.(\w+)\]/gm,
	seqDef: /^\[(lane|riff)\.(\w+)\]/gm,
};

/** 
 * @param {string} name
 * @param {string} source
 */
function getSequenceRaw(name, source) {
	const regex = new RegExp(`^\\[(lane|riff)\\.${name}\\]`, 'm');
	const match = regex.exec(source);
	if (!match) return null;
	const start = match.index + match[0].length;
	const nextIdx = source.indexOf('\n[', start);
	const end = nextIdx === -1 ? source.length : nextIdx;
	return source.substring(start, end).trim();
}

/**
 * @param {string} raw
 */
function getSequenceDetails(raw) {
	const [header, ...rest] = raw.trim().split('\n');
	// const remaining = rest.join('\n');
	const [, type, name] = regExs.seqDef.exec(header) || [];
	// const [, vars = ''] = remaining.match(/vars\s*=\s*['"](.+?)['"]/) || [];
	// const [, tick = ''] = remaining.match(/tick\s*=\s*['"](.+?)['"]/) || [];
	// const [, tune = ''] = remaining.match(/tune\s*=\s*['"](.+?)['"]/) || [];
	// const [, content = ''] = remaining.match(/content\s*=\s*'''([\s\S]*?)'''/) || [];
	return {
		type,
		name,
		// vars: vars.trim(),
		// tick: tick.trim(),
		// tune: tune.trim(),
		// content: content.trim(),
	};
}

/**
 * @param {number} ms
 */
function sleep(ms = 0) {
	return new Promise(r => setTimeout(r, ms))
}

module.exports = {
	sleep,
	regExs,
	parseBleedData,
	getSequenceHeader,
	getSequenceRaw,
	getSequenceDetails,
};
