import {loadCardImages, onLoaded, loadImage, Button, Card, Hand, Board, Stack} from './ui.js';
import {$, $$} from './util.js';

window.addEventListener('load', () => {
	loadCardImages();

	const canvas = $('#board');
	const board = new Board({canvas});
	const swordImg = loadImage('/images/sword.png');
	const shieldImg = loadImage('/images/shield.png');
	const names = ['Emote', 'Artichoke', 'Json', 'Wrench', 'Ziggler', 'Fanny'];
	const offsets = [-200, 200, -200, 200, 0, 0];

	new Stack({board, cards: [new Card('hearts', '2')]});
	new Stack({board, cards: [new Card('hearts', '2'), new Card('hearts', '9')]});
	new Stack({board, cards: [new Card('diamonds', 'jack'), new Card('hearts', '9'), new Card('spades', 'ace')]});

	['top', 'top', 'bottom', 'bottom', 'left', 'right'].forEach((lrtb, i) => {
		const hand = new Hand({board, lrtb, name: names[i], offset: offsets[i]});

		hand.cards.push(new Card('hearts', '3'));
		hand.cards.push(new Card('hearts', '3'));
		hand.cards.push(new Card('hearts', '3'));
		hand.cards.push(new Card('hearts', '3'));
		hand.cards.push(new Card('hearts', '3'));
		hand.cards.push(new Card('hearts', '3'));
		hand.cards.push(new Card('hearts', '3'));
		hand.cards.push(new Card('hearts', '3'));
		hand.cards.push(new Card('hearts', '3'));
		hand.cards.push(new Card('hearts', '3'));
		hand.cards.push(new Card('hearts', '3'));
		hand.cards.push(new Card('hearts', '3'));
		hand.cards.push(new Card('hearts', '3'));
		hand.cards.at(-2).selected = true;

		if (i == 1) {
			hand.buttons.push(new Button({img: swordImg}));
		} else if (i == 2) {
			hand.buttons.push(new Button({img: shieldImg}));
		}
		hand.buttons.push(new Button({text: 'Picking Up'}));
		hand.buttons.push(new Button({text: 'Okay', cb: true}));
		hand.buttons.push(new Button({text: 'Pass', cb: true}));
	});

	canvas.addEventListener('mousemove', e => {
		board.mousemove(e);
	});
	
	canvas.addEventListener('mouseout', e => {
		board.mouseout();
	});

	canvas.addEventListener('mousedown', e => {
		board.mousedown(e);
	});

	canvas.addEventListener('mouseup', e => {
		board.mouseup();
	});

	onLoaded(() => board.draw());
});
