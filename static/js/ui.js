export {loadCardImages, onLoaded, Card, Hand, Board};

const cardImages = {};
let cardImagesLoaded = 0;
const suits = ['hearts', 'diamonds', 'clubs', 'spades'];
const ranks = ['2', '3', '4', '5', '6', '7', '8', '9', '10', 'jack', 'queen', 'king', 'ace'];

function loadCardImages() {
	suits.forEach(s => {
		ranks.forEach(r => {
			const name = s + '_' + r;
			const img = new Image();
			img.src = `/cards/fronts/${name}.png`;
			img.addEventListener('load', () => {
				cardImagesLoaded++;
			});
			cardImages[name] = img;
		});
	});
	const img = new Image();
	img.src = '/cards/backs/astronaut.png';
	img.addEventListener('load', () => {
		cardImagesLoaded++;
	});
	cardImages['back'] = img;
}

function onLoaded(fn) {
	setTimeout(() => {
		if (cardImagesLoaded == 53) {
			fn();
		} else {
			onLoaded(fn);
		}
	}, 10);
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
		this.width = params.width ?? 250;
		this.defSep = params.defSep ?? 30;
		this.cards = [];
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
			const ori = this.width/2-this.cards[0].width/2;
			const cori = ori - 0.5*sep*(n-1) + sep*i;
			this.ctx.save();
			this.ctx.translate(cori, 0);
			if (this.lrtb == 'bottom') {
				const dx = this.canvas.width/2-this.width/2;
				const dy = this.canvas.height-2*this.cards[0].height/3;
				this.ctx.translate(dx, dy);
				this.ctx.rotate(-0.02*(n-1) + 0.04*i);
				this.ctx.translate(0, Math.abs(-0.25*(n-1) + 0.5*i)**2);
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
	}
}
