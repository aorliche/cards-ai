import {loadCardImages, onLoaded, loadImage, Button, Card, Hand, Board, Stack} from './ui.js';
import {$, $$} from './util.js';

window.addEventListener('load', () => {
	loadCardImages();

	const suits = ["clubs", "spades", "hearts", "diamonds"];
	const ranks = ["6", "7", "8", "9", "10", "jack", "queen", "king", "ace"];

	const names = ['Emote', 'Artichoke', 'Json', 'Wrench', 'Ziggler', 'Fanny'];
	let name = names[Math.floor(names.length*Math.random())];
	$('#name').value = name;

	let players = ['Human'];
	let gameId = -1;
	let conn = null;

	// Update the list of games
	function updateList(lst) {
		const existing = $$('#games-select option');
		const myGameName = `Game ${gameId}`;
		// Remove invalid existing entries
		for (let i=0; i<existing.length; i++) {
			let found = false;
			for (let j=0; j<lst.length; j++) {
				const name = `Game ${lst[j]}`;
				if (name == existing[i].innerText) {
					found = true;
					break;
				}
			}
			if (!found || existing[i].innerText == myGameName) {
				existing[i].parentNode.removeChild(existing[i]);
			}
		}
		// Add new entries
		for (let i=0; i<lst.length; i++) {
			const name = `Game ${lst[i]}`;
			if (name == myGameName) {
				continue;
			}
			let found = false;
			for (let i=0; i<existing.length; i++) {
				if (existing[i].innerText == name) {
					found = true;
					break;
				}
			}
			if (!found) {
				const opt = document.createElement('option');
				opt.innerText = name;
				$('#games-select').appendChild(opt);
			}
		}
	}

	const canvas = $('#board');
	const board = new Board({canvas});
	const swordImg = loadImage('/images/sword.png');
	const shieldImg = loadImage('/images/shield.png');

	// Dummy hand
	function makeDummyHand() {
		board.hands = [];
		const hand = new Hand({board, lrtb: 'bottom', name, offset: 0});

		for (let i=0; i<6; i++) {
			hand.cards.push(new Card('hearts', '2'));
			hand.cards.at(-1).visible = false;
		}
	}

	makeDummyHand();

	// Update board
	// Go clockwise from bottom
	function updateBoard(data) {
		const p = data.Player;
		const nh = data.State.Hands.length;
		let lrtbs = ['bottom', 'top'];
		let offsets = [0, 0];
		if (nh == 3) {
			lrtbs = ['bottom', 'top', 'top'];
			offsets = [0, -200, 200];
		} else if (nh == 4) {
			lrtbs = ['bottom', 'left', 'top', 'right'];
			offsets = [0, 0, 0, 0];
		}
		// Rebuild hands if necessary
		if (board.hands.length != nh) {
			board.hands = [];

			for (let i=0; i<nh; i++) {
				const j = (p+i)%nh;
				const hand = new Hand({board, lrtb: lrtbs[i], name: data.Names[j], offset: offsets[i]});
			}
		}
		for (let i=0; i<nh; i++) {
			const j = (p+i)%nh;
			const hand = board.hands[i];

			// Update names as players join
			hand.name = data.Names[j];
			
			// Ditch holding
			if (hand.holding) {
				hand.holding = null;
			}
			
			// Update cards
			hand.cards = [];
			for (let k=0; k<data.State.Hands[j].length; k++) {
				const cardIdx = data.State.Hands[j][k];
				let suit, rank, card;
				if (cardIdx == -1) {
					suit = suits[0];
					rank = ranks[0];
					card = new Card(suit, rank);
					card.visible = false;
				} else {
					suit = suits[Math.floor(cardIdx/9)];
					rank = ranks[cardIdx % 9];
					console.log(suit, rank);
					card = new Card(suit, rank);
					card.visible = true;
				}
				// Don't show even known enemy cards
				if (i != 0) {
					card.visible = false;
				}
				hand.cards.push(card);
			}
		}
		board.draw();
	}

	conn = new WebSocket(`ws://${location.host}/ws`);

	/*conn.onopen = () => {
		conn.send(JSON.stringify({'Type': 'List'}));
	}*/

	conn.onmessage = e => {
		const json = JSON.parse(e.data);
		console.log(json);
		const data = json.Data ? JSON.parse(json.Data) : null;
		switch (json.Type) {
			case 'Error':
				break;
			case 'List':
				updateList(data);
				break;
			case 'New':
				gameId = data;
				break;
			case 'Join':
				gameId = data;
				break;
			case 'Update':
				updateBoard(data);
				break;
			case 'Chat':
				$('#chat').value += `${data.Name}: ${data.Message}\n`;
				$('#chat').scrollTop = $('#chat').scrollHeight;
				break;
		}
	}

	setInterval(() => {
		if (conn.readyState === WebSocket.OPEN) {
			conn.send(JSON.stringify({'Type': 'List'}));
		}
	}, 1000);

	$('#start').addEventListener('click', () => {
		makeDummyHand();
		conn.send(JSON.stringify({'Type': 'New', 'Types': players, 'Name': $('#name').value}));
	});

	$('#join').addEventListener('click', () => {
		makeDummyHand();
		const select = $('#games-select');
		const opt = select.options[select.selectedIndex];
		if (!opt) {
			return;
		}
		const key = parseInt(opt.innerText.slice(5));
		conn.send(JSON.stringify({'Type': 'Join', 'Game': key, 'Name': $('#name').value}));
	});

	function sendChat() {
		conn.send(JSON.stringify({'Type': 'Chat', 'Game': gameId, 'Data': $('#message').value}));
		$('#message').value = "";
	}

	$('#send').addEventListener('click', sendChat);
	$('#message').addEventListener('keyup', e => {
		if (e.code == 'Enter') {
			sendChat();
		}
	});

	/*new Stack({board, cards: [new Card('hearts', '2')]});
	new Stack({board, cards: [new Card('hearts', '2'), new Card('hearts', '9')]});
	new Stack({board, cards: [new Card('diamonds', 'jack'), new Card('hearts', '9'), new Card('spades', 'ace')]});*/

	/*['top', 'top', 'bottom', 'bottom', 'left', 'right'].forEach((lrtb, i) => {
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
	});*/

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

	function rebuildPlayers() {
		const div = $('#players-inner');
		div.innerHTML = '';
		for (let i=0; i<players.length; i++) {
			const odiv = document.createElement('div');
			const pdiv = document.createElement('div');
			const a = document.createElement('a');
			pdiv.classList.add('type');
			pdiv.classList.add(players[i]);
			pdiv.innerHTML = players[i];
			a.href = '#';
			a.addEventListener('click', e => {
				e.preventDefault();
				players.splice(i, 1);
				rebuildPlayers();
			});
			a.innerText = 'x';
			odiv.appendChild(pdiv);
			odiv.appendChild(a);
			div.appendChild(odiv);
		}
		$('#number').innerText = `${players.length} Players`;
	}

	function addPlayer(type) {
		players.push(type);
		rebuildPlayers();
	}

	$('#human').addEventListener('click', e => {
		addPlayer('Human');
	});
	
	$('#easy').addEventListener('click', e => {
		addPlayer('Easy');
	});
	
	$('#medium').addEventListener('click', e => {
		addPlayer('Medium');
	});
});
