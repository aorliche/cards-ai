import {loadCardImages, onLoaded, Card, Hand, Board} from './ui.js';
import {$, $$} from './util.js';

window.addEventListener('load', () => {
	loadCardImages();

	const canvas = $('#board');
	const board = new Board({canvas});
	const hand = new Hand({board});

	hand.cards.push(new Card('hearts', '2'));
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

	canvas.addEventListener('mousemove', e => {
		board.mousemove(e);
	});
	
	canvas.addEventListener('mouseout', e => {
		board.mouseout();
	});

	onLoaded(() => hand.draw());
});
