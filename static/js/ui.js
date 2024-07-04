export {loadCardImages, onLoaded, Button, Card, Hand, Board};

const cardImages = {};
const buttonImages = {};
let imagesLoaded = 0;
const suits = ['hearts', 'diamonds', 'clubs', 'spades'];
const ranks = ['2', '3', '4', '5', '6', '7', '8', '9', '10', 'jack', 'queen', 'king', 'ace'];
const buttons = ['pass_text', 'pick_up_text', 'pass', 'pass_over', 'okay', 'okay_over', 'pick_up', 'pick_up_over'];

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
	buttons.forEach(name => {
		const img = new Image();
		img.src = `/images/buttons/${name}.png`;
		img.addEventListener('load', () => {
			imagesLoaded++;
		});
		buttonImages[name] = img;
	});
}

function totalImages() {
	return suits.length*ranks.length + 1 + buttons.length;
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

class Button {
	constructor(name, cb) {
		this.name_ = name;
		this.cb = cb;
	}

	get width() {
		return buttonImages[this.name].width;
	}

	get height() {
		return buttonImages[this.name].height;
	}

	get name() {
		if (this.hovering) {
			const n = this.name_ + '_over';
			if (buttons.includes(n)) {
				return n;
			}
		}
		return this.name_;
	}
	
	draw(ctx) {
		ctx.drawImage(buttonImages[this.name], 0, 0, this.width, this.height);
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

class Board {
	constructor(params) {
		this.canvas = params.canvas;
		if (!this.canvas) {
			throw Error('Board constructor not passed canvas param');
		}
		this.ctx = this.canvas.getContext('2d');
		this.hands = [];
	}

	draw() {
		this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
		this.hands.forEach(h => h.draw());
	}

	mousemove(e) {
		this.mouseout();
		this.hands.forEach(h => {
			h.mousemove(e);
		});
		this.draw();
	}

	mouseout() {
		this.hands.forEach(h => {
			h.cards.forEach(c => {
				c.hovering = false;
			});
			h.buttons.forEach(b => {
				b.hovering = false;
			});
		});
		this.draw();
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
		this.cards = [];
		this.buttons = [];
	}

	getButtonTransforms() {
		const xfms = [];
		if (this.lrtb == 'bottom') {
			const ori = this.canvas.width/2;
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
				const ch = this.cards.length > 0 ? this.cards[0].height + 50 : 50;
				const dy = this.canvas.height - ch + (h - this.buttons[i].height)/2;
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
				const dx = this.canvas.width/2-this.width/2;
				const dy = this.canvas.height-2*this.cards[0].height/3;
				this.ctx.translate(dx, dy);
				this.ctx.rotate(-0.02*(n-1) + 0.04*i);
				this.ctx.translate(0, 2*Math.abs(-0.25*(n-1) + 0.5*i)**2);
				xfms.push(this.ctx.getTransform());
			} else if (this.lrtb == 'top') {
				const ori = this.width/2+this.cards[0].width/2;
				const cori = ori + 0.5*sep*(n-1) - sep*i;
				this.ctx.translate(cori, 0);
				const dx = this.canvas.width/2-this.width/2;
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
	}
}
