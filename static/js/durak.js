import {loadCardImages, onLoaded, Button, Card, Hand, Board} from './ui.js';
import {$, $$} from './util.js';

window.addEventListener('load', () => {
	loadCardImages();

	const canvas = $('#board');
	const board = new Board({canvas});
	const names = ['Emote', 'Wrench', 'Ziggler', 'Fanny'];

	['top', 'bottom', 'left', 'right'].forEach((lrtb, i) => {
		const hand = new Hand({board, lrtb, name: names[i]});

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
			hand.buttons.push(new Button({text: 'Attacking'}));
		} else if (i == 2) {
			hand.buttons.push(new Button({text: 'Defending'}));
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
