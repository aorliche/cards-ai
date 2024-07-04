import {loadCardImages, onLoaded, Button, Card, Hand, Board} from './ui.js';
import {$, $$} from './util.js';

window.addEventListener('load', () => {
	loadCardImages();

	const canvas = $('#board');
	const board = new Board({canvas});

	['top', 'bottom', 'left', 'right'].forEach(lrtb => {
		const hand = new Hand({board, lrtb});

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

		hand.buttons.push(new Button('pick_up_text'));
		hand.buttons.push(new Button('okay'));
	});

	canvas.addEventListener('mousemove', e => {
		board.mousemove(e);
	});
	
	canvas.addEventListener('mouseout', e => {
		board.mouseout();
	});

	onLoaded(() => board.draw());
});
