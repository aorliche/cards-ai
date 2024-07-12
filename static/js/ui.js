export {loadCardImages, onLoaded, loadImage, Button, Stack, Card, Hand, Board, Deck};

import {drawText} from './util.js';

const cardImages = {};
let imagesLoaded = 0;
const suits = ['hearts', 'diamonds', 'clubs', 'spades'];
const ranks = ['2', '3', '4', '5', '6', '7', '8', '9', '10', 'jack', 'queen', 'king', 'ace'];

function loadImage(src) {
	const img = new Image();
	img.src = src;
	return img;
}

function loadCardImages() {
	suits.forEach(s => {
		ranks.forEach(r => {
			const name = s + '_' + r;
			const img = new Image();
			img.src = `/cards/fronts/${name}.png`;
			img.addEventListener('load', () => {
				imagesLoaded++;
			});
			cardImages[name] = img;
		});
	});
	const img = new Image();
	img.src = '/cards/backs/astronaut.png';
	img.addEventListener('load', () => {
		imagesLoaded++;
	});
	cardImages['back'] = img;
}

function totalImages() {
	return suits.length*ranks.length + 1;
}

function onLoaded(fn) {
	setTimeout(() => {
		if (imagesLoaded == totalImages()) {
			fn();
		} else {
			onLoaded(fn);
		}
	}, 10);
}

// So you don't have to pass ctx to button constructor
const buttonCanvas = document.createElement('canvas');

class Button {
	constructor(params) {
		this.img = params.img ?? null;
		this.text = params.text ?? 'Button';
		this.cb = params.cb ?? null;
		this.font = params.font ?? (this.cb ? '18px sans' : '16px sans');
		this.padding = params.padding ?? 20;
		// Measure text width and height
		const ctx = buttonCanvas.getContext('2d');
		ctx.save();
		ctx.font = this.font;
		this.tm = ctx.measureText(this.text);
		ctx.restore();
	}

	get width() {
		if (this.img) {
			return this.img.width;
		}
		const xt = this.tm.width;
		const xb = this.cb ? this.padding : 0;
		return xt + xb;
	}

	get height() {
		if (this.img) {
			return this.img.height;
		}
		if (this.cb) {
			return 40;
		}
		const yp = this.tm.actualBoundingBoxAscent;
		const ym = this.tm.actualBoundingBoxDescent;
		return yp + ym;
	}
	
	draw(ctx) {
		if (this.img) {
			ctx.save();
			ctx.drawImage(this.img, 0, 0);
			ctx.restore();
		} else if (this.cb) {
			ctx.save();
			ctx.lineWidth = 3;
			ctx.strokeStyle = '#a2a';
			ctx.fillStyle = this.hovering ? '#f6f' : '#77f';
			ctx.font = 'bold ' + this.font;
			ctx.beginPath();
			ctx.roundRect(0, 0, this.width, this.height, 3*this.padding/4);
			ctx.stroke();
			ctx.beginPath();
			ctx.roundRect(1, 1, this.width-2, this.height-2, 3*this.padding/4);
			ctx.fill();
			ctx.fillStyle = '#fff';
			const yp = this.tm.actualBoundingBoxAscent;
			const ym = this.tm.actualBoundingBoxDescent;
			const diff = this.height - (yp + this.padding);
			ctx.fillText(this.text, this.padding/2, this.padding/2+yp+diff/2-2);
			ctx.restore();
		} else {
			ctx.save();
			ctx.fillStyle = '#000';
			ctx.font = this.font;
			ctx.fillText(this.text, 0, this.tm.actualBoundingBoxAscent);
			ctx.restore();
		}
	}
}

class Card {
	constructor(suit, rank, visible) {
		this.suit = suit;
		this.rank = rank;
		this.visible = visible ?? true;
		this.width = 80;
		this.hovering = false;
	}

	get height() {
		return this.trueHeight*this.width/this.trueWidth
	}

	get trueWidth() {
		if (!cardImages[this.name]) {
			return 100;
		}
		return cardImages[this.name].width;
	}

	get trueHeight() {
		if (!cardImages[this.name]) {
			return 100;
		}
		return cardImages[this.name].height;
	}

	get name() {
		if (!this.visible) {
			return 'back';
		}
		return this.suit + '_' + this.rank;
	}

	draw(ctx) {
		if (!cardImages[this.name]) {
			return;
		}
		const dy = this.hovering ? -20 : 0;
		if (this.selected) {
			ctx.fillStyle = '#f00';
			ctx.fillRect(-2, dy-2, this.width+4, this.height+4);
		}
		ctx.drawImage(cardImages[this.name], 0, dy, this.width, this.height);
	}
}

class Stack {
	constructor(params) {
		this.board = params.board;
		if (!this.board) {
			throw Error('Stack constructor not passed board');
		}
		this.board.stacks.push(this);
		this.ctx = this.board.ctx;
		this.cards = params.cards ?? [];
		this.dy = 20;
		this.dx = 10;
	}

	get stackIndex() {
		for (let i=0; i<this.board.stacks.length; i++) {
			if (this.board.stacks[i] == this) return i;
		}
		return -1;
	}

	getCardTransforms() {
		const sxfm = this.board.getStackTransforms()[this.stackIndex];
		if (!sxfm) {
			return null;
		}
		const xfms = [];
		for (let i=0; i<this.cards.length; i++) {
			this.ctx.save();
			this.ctx.setTransform(sxfm);
			this.ctx.translate(i*this.dx, i*this.dy);
			xfms.push(this.ctx.getTransform());
			this.ctx.restore();
		}
		return xfms;
	}

	draw() {
		const xfms = this.getCardTransforms();
		if (xfms) {
			for (let i=0; i<xfms.length; i++) {
				this.ctx.save();
				this.ctx.setTransform(xfms[i]);
				this.cards[i].draw(this.ctx);
				this.ctx.restore();
			}
		}
	}
}

class Deck {
	constructor(params) {
		this.over = new Card('hearts', '2');
		this.over.visible = false;
		this.trump = params.trump ?? null;
		this.p = params.p ?? {x: 100, y: 60};
		this.size = params.size ?? 100;
	}

	draw(ctx) {
		if (this.size > 0) {
			ctx.save();
			ctx.translate(this.p.x, this.p.y+40);
			ctx.translate(-this.over.width/2, -this.over.height/2);
			this.trump.draw(ctx)
			ctx.restore();
		}
		if (this.size > 1) {
			ctx.save();
			ctx.translate(this.p.x, this.p.y);
			ctx.save();
			ctx.translate(this.over.height/2, -this.over.width/2);
			ctx.rotate(Math.PI/2);
			this.over.draw(ctx)
			ctx.restore();
			drawText(ctx, this.size, {x: 0, y: 20}, 'red', 'bold 48px sans');
			ctx.restore();
		}
	}
}

class Board {
	constructor(params) {
		this.canvas = params.canvas;
		if (!this.canvas) {
			throw Error('Board constructor not passed canvas param');
		}
		this.ctx = this.canvas.getContext('2d');
		this.hands = [];
		this.stacks = [];
		this.message_ = null;
		this.deck = null;
	}

	getStackTransforms() {
		const xfms = [];
		const n = this.stacks.length; 
		let w = 0;
		let h = 0;
		const p = [0];
		for (let i=0; i<n; i++) {
			if (this.stacks[i].cards.length > 0) {
				w = this.stacks[i].cards[0].width;
				h = this.stacks[i].cards[0].height;
				break;
			}
		}
		for (let i=0; i<n; i++) {
			p.push(p.at(-1) + w + 20);
		}
		const dx = this.canvas.width/2 - p.at(-1)/2;
		const dy = this.canvas.height/2 - h/2;
		for (let i=0; i<n; i++) {
			this.ctx.save();
			this.ctx.translate(dx + p[i], dy);
			xfms.push(this.ctx.getTransform());
			this.ctx.restore();
		}
		return xfms;
	}

	draw() {
		this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
		if (this.deck) {
			this.deck.draw(this.ctx);
		}
		this.stacks.forEach(s => s.draw());
		this.hands.forEach(h => h.draw());
		if (this.message_) {
			drawText(this.ctx, this.message_, {x: this.canvas.width/2, y: this.canvas.height/2+20}, 'red', 'bold 64px sans');
		}
	}

	mousedown(e) {
		this.hands.forEach(h => {
			h.mousedown(e);
			h.mousemove(e);
		});
		this.draw();
	}

	mousemove(e) {
		this.hands.forEach(h => {
			h.cards.forEach(c => c.hovering = false);
			h.buttons.forEach(b => b.hovering = false);
			h.mousemove(e);
		});
		this.draw();
	}

	mouseout() {
		this.hands.forEach(h => {
			h.mouseout();
		});
		this.draw();
	}

	mouseup() {
		this.hands.forEach(h => {
			h.mouseup();
		});
		this.draw();
	}

	get message() {
		return this.message_;
	}

	set message(msg) {
		this.message_ = msg;
	}
}

class Hand {
	constructor(params) {
		this.board = params.board;
		if (!this.board) {
			throw Error('Hand constructor not passed board');
		}
		this.board.hands.push(this);
		this.canvas = this.board.canvas;
		this.ctx = this.board.ctx;
		// left, right, top, bottom
		this.lrtb = params.lrtb ?? 'bottom';
		this.offset = params.offset ?? 0;
		this.width = params.width ?? 300;
		this.defSep = params.defSep ?? 30;
		this.name = params.name ?? 'Unnamed';
		this.cards = [];
		this.buttons = [];
		this.holding = null;
	}

	getButtonTransforms() {
		const xfms = [];
		if (this.lrtb == 'bottom') {
			const ori = this.offset + this.canvas.width/2;
			let w = [0];
			let h = 0;
			for (let i=0; i<this.buttons.length; i++) {
				w.push(w.at(-1) + this.buttons[i].width + 10);
				if (this.buttons[i].height > h) {
					h = this.buttons[i].height;
				}
			}
			for (let i=0; i<this.buttons.length; i++) {
				w[i] -= w.at(-1)/2;
				const dx = ori + w[i];
				const ch = this.cards.length > 0 ? 2*this.cards[0].height/3 + 60 : 60;
				const dy = this.canvas.height - ch + (h - this.buttons[i].height)/2;
				this.ctx.save();
				this.ctx.translate(dx, dy); 
				xfms.push(this.ctx.getTransform());
				this.ctx.restore();
			}
		} else if (this.lrtb == 'top') {
			const ori = this.offset + this.canvas.width/2;
			let w = [0];
			let h = 0;
			for (let i=0; i<this.buttons.length; i++) {
				w.push(w.at(-1) + this.buttons[i].width + 10);
				if (this.buttons[i].height > h) {
					h = this.buttons[i].height;
				}
			}
			for (let i=0; i<this.buttons.length; i++) {
				w[i] -= w.at(-1)/2;
				const dx = ori + w[i];
				const ch = this.cards.length > 0 ? 2*this.cards[0].height/3 + 20 : 20;
				const dy = ch + (h - this.buttons[i].height)/2;
				this.ctx.save();
				this.ctx.translate(dx, dy); 
				xfms.push(this.ctx.getTransform());
				this.ctx.restore();
			}
		} else if (this.lrtb == 'left') {
			const ori = this.canvas.height/2;
			let h = [0];
			for (let i=0; i<this.buttons.length; i++) {
				h.push(h.at(-1) + this.buttons[i].height + 10);
			}
			for (let i=0; i<this.buttons.length; i++) {
				h[i] -= h.at(-1)/2;
				const dy = ori + h[i];
				const ch = this.cards.length > 0 ? 2*this.cards[0].height/3 + 20 : 20;
				const dx = ch;
				this.ctx.save();
				this.ctx.translate(dx, dy); 
				xfms.push(this.ctx.getTransform());
				this.ctx.restore();
			}
		} else if (this.lrtb == 'right') {
			const ori = this.canvas.height/2;
			let h = [0];
			let w = 0;
			for (let i=0; i<this.buttons.length; i++) {
				h.push(h.at(-1) + this.buttons[i].height + 10);
				if (this.buttons[i].width > w) {
					w = this.buttons[i].width;
				}
			}
			for (let i=0; i<this.buttons.length; i++) {
				h[i] -= h.at(-1)/2;
				const dy = ori + h[i];
				const ch = this.cards.length > 0 ? 2*this.cards[0].height/3 + 20 : 20;
				const dx = this.canvas.width - ch - w;
				this.ctx.save();
				this.ctx.translate(dx, dy); 
				xfms.push(this.ctx.getTransform());
				this.ctx.restore();
			}
		}
		return xfms;
	}

	getCardTransforms() {
		let sep;
		const n = this.cards.length;
		if (n == 0) {
			return;
		} else if (n == 1) {
			sep = 0;
		} else {
			sep = (this.width - this.cards[0].width)/(n-1);
			if (sep > this.defSep) {
				sep = this.defSep;
			}
		}
		const xfms = [];
		for (let i=0; i<n; i++) {
			this.ctx.save();
			if (this.lrtb == 'bottom') {
				const ori = this.width/2-this.cards[0].width/2;
				const cori = ori - 0.5*sep*(n-1) + sep*i;
				this.ctx.translate(cori, 0);
				const dx = this.offset + this.canvas.width/2-this.width/2;
				const dy = this.canvas.height-2*this.cards[0].height/3;
				this.ctx.translate(dx, dy);
				this.ctx.rotate(-0.02*(n-1) + 0.04*i);
				this.ctx.translate(0, 2*Math.abs(-0.25*(n-1) + 0.5*i)**2);
				xfms.push(this.ctx.getTransform());
			} else if (this.lrtb == 'top') {
				const ori = this.width/2+this.cards[0].width/2;
				const cori = ori + 0.5*sep*(n-1) - sep*i;
				this.ctx.translate(cori, 0);
				const dx = this.offset + this.canvas.width/2-this.width/2;
				const dy = 2*this.cards[0].height/3;
				this.ctx.translate(dx, dy);
				this.ctx.rotate(3.14);
				this.ctx.rotate(-0.02*(n-1) + 0.04*i);
				this.ctx.translate(0, 2*Math.abs(0.25*(n-1) - 0.5*i)**2);
				xfms.push(this.ctx.getTransform());
			} else if (this.lrtb == 'left') {
				const ori = this.width/2-this.cards[0].width/2;
				const cori = ori - 0.5*sep*(n-1) + sep*i;
				this.ctx.translate(0, cori);
				const dy = this.canvas.height/2-this.width/2;
				const dx = 2*this.cards[0].height/3;
				this.ctx.translate(dx, dy);
				this.ctx.rotate(3.14/2);
				this.ctx.rotate(-0.02*(n-1) + 0.04*i);
				this.ctx.translate(0, 2*Math.abs(0.25*(n-1) - 0.5*i)**2);
				xfms.push(this.ctx.getTransform());
			} else if (this.lrtb == 'right') {
				const ori = this.width/2+this.cards[0].width/2;
				const cori = ori + 0.5*sep*(n-1) - sep*i;
				this.ctx.translate(0, cori);
				const dy = this.canvas.height/2-this.width/2;
				const dx = this.canvas.width-2*this.cards[0].height/3;
				this.ctx.translate(dx, dy);
				this.ctx.rotate(-3.14/2);
				this.ctx.rotate(-0.02*(n-1) + 0.04*i);
				this.ctx.translate(0, 2*Math.abs(0.25*(n-1) - 0.5*i)**2);
				xfms.push(this.ctx.getTransform());
			}
			this.ctx.restore();
		}
		return xfms;
	}

	draw() {
		const xfms = this.getCardTransforms();
		if (xfms) {
			for (let i=0; i<xfms.length; i++) {
				this.ctx.save();
				this.ctx.setTransform(xfms[i]);
				this.cards[i].draw(this.ctx);
				this.ctx.restore();
			}
		}
		const bxfms = this.getButtonTransforms();
		if (bxfms) {
			for (let i=0; i<bxfms.length; i++) {
				this.ctx.save();
				this.ctx.setTransform(bxfms[i]);
				this.buttons[i].draw(this.ctx);
				this.ctx.restore();
			}
		}
		if (this.holding) {
			this.ctx.save();
			this.ctx.translate(this.holdingPos.x-this.holding.width/2, this.holdingPos.y-this.holding.height/2);
			this.holding.draw(this.ctx);
			this.ctx.restore();
		}
		if (this.lrtb == 'bottom') {
			drawText(this.ctx, this.name, {x: this.offset + this.canvas.width/2 - this.width/2 - 20, y: this.canvas.height - 10}, 'black', '16px sans');
		} else if (this.lrtb == 'top') {
			drawText(this.ctx, this.name, {x: this.offset + this.canvas.width/2 - this.width/2 - 20, y: 25}, 'black', '16px sans');
		} else if (this.lrtb == 'left') {
			drawText(this.ctx, this.name, {x: 40, y: this.canvas.height/2 - this.width/2 - 10}, 'black', '16px sans');
		} else if (this.lrtb == 'right') {
			drawText(this.ctx, this.name, {x: this.canvas.width - 40, y: this.canvas.height/2 - this.width/2 - 10}, 'black', '16px sans');
		}
	}

	mousemove(e) {
		const xfms = this.getCardTransforms();
		const point = new DOMPoint(e.offsetX, e.offsetY);
		if (xfms) {
			for (let i=xfms.length-1; i>=0; i--) {
				const p = point.matrixTransform(xfms[i].inverse());
				if (p.x > 0 && p.x < this.cards[0].width && p.y > 0 && p.y < this.cards[0].height) {
					this.cards[i].hovering = true;
					break;
				}
			}
		}
		const bxfms = this.getButtonTransforms();
		if (bxfms) {
			for (let i=bxfms.length-1; i>=0; i--) {
				const p = point.matrixTransform(bxfms[i].inverse());
				if (p.x > 0 && p.x < this.buttons[i].width && p.y > 0 && p.y < this.buttons[i].height) {
					this.buttons[i].hovering = true;
					break;
				}
			}
		}
		if (this.holding) {
			this.holdingPos = point;
		}
	}

	mousedown(e) {
		for (let i=0; i<this.cards.length; i++) {
			if (this.cards[i].hovering) {
				this.cards[i].hovering = false;
				this.holding = this.cards[i];
				this.holdingPos = new DOMPoint(e.offsetX, e.offsetY);
				this.cards.splice(i, 1);
				break;
			}
		}
	}

	mouseout() {
		this.cards.forEach(c => {
			c.hovering = false;
		});
		this.buttons.forEach(b => {
			b.hovering = false;
		});
		if (this.holding) {
			this.cards.push(this.holding);
			this.holding = null;
		}
	}

	mouseup() {
		if (this.holding) {
			this.cards.push(this.holding);
			this.holding = null;
		}
		for (let i=0; i<this.buttons.length; i++) {
			if (this.buttons[i].hovering && this.buttons[i].cb) {
				this.buttons[i].cb();
			}
		}
	}
}
