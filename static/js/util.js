export {$, $$, drawText};

const $ = q => document.querySelector(q);
const $$ = q => [...document.querySelectorAll(q)];

function drawText(ctx, text, p, color, font, stroke) {
	ctx.save();
	if (font) ctx.font = font;
	const tm = ctx.measureText(text);
	ctx.fillStyle = color;
	if (p.ljust) 
		ctx.fillText(text, p.x, p.y);
	else if (p.rjust)
		ctx.fillText(text, p.x-tm.width, p.y);
	else
		ctx.fillText(text, p.x-tm.width/2, p.y);
	if (stroke) {
		ctx.strokeStyle = stroke;
		ctx.lineWidth = 1;
		ctx.strokeText(text, p.x-tm.width/2, p.y);
	}
	ctx.restore();
	return tm;
}
