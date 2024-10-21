import {loadCardImages, onLoaded, loadImage, Button, Card, Hand, Board, CircleStack, Deck} from './ui.js';
import {$, $$} from './util.js';

window.addEventListener('load', () => {
	loadCardImages();
	
	const suits = ["clubs", "spades", "hearts", "diamonds"];
	const ranks = ["2", "3", "4", "5", "6", "7", "8", "9", "10", "jack", "queen", "king", "ace"];
	const verbs = ["Bid", "Play"];

	function cardToIndex(card) {
		let i = suits.indexOf(card.suit);
		let j = ranks.indexOf(card.rank);
		if (i != -1 && j != -1) {
			return i*13+j;
		}
		return null;
	}

	const names = ['Emote', 'Artichoke', 'Json', 'Wrench', 'Ziggler', 'Jobber'];
	let name = names[Math.floor(names.length*Math.random())];
	$('#name').value = name;

	let players = ['Human'];
	let gameId = -1;
	let playerId = -1;
	let myActions = [];
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
	
	// Dummy hand
	function makeDummyHand() {
		board.hands = [];

		['bottom', 'left', 'top', 'right'].forEach(val => {
			const hand = new Hand({board, lrtb: val, name, offset: 0});

			for (let i=0; i<6; i++) {
				hand.cards.push(new Card('hearts', '2'));
				hand.cards.at(-1).visible = true;
			}
		});

		board.stacks = [];
		const stack1 = new CircleStack({
			board, 
			cards: [
				new Card('hearts', '3'), 
				new Card('spades', 'ace'),
				new Card('clubs', 'king'),
				new Card('diamonds', '8'),
			],
			start: 1
		});

		for (let i=0; i<board.hands.length; i++) {
			const hand = board.hands[i];
			let bidv = 0;
			const bidb = new Button({text: 'Bid: ' + bidv});

			hand.buttons = [];
			hand.buttons.push(bidb);

			if (i == 0) {
				hand.buttons.push(new Button({
					text: '-',
					cb: () => {
						if (bidv > 0) {
							bidv--;
							bidb.text = 'Bid: ' + bidv;
							board.draw();
						}
					},
				}));
				hand.buttons.push(new Button({
					text: '+',
					cb: () => {
						if (bidv < 13) {
							bidv++;
							bidb.text = 'Bid: ' + bidv;
							board.draw();
						}
					},
				}));
				hand.buttons.push(new Button({
					text: 'Place Bid',
					cb: () => {
						hand.buttons = [];
						hand.buttons.push(new Button({text: 'Bid: ' + bidv}));
						board.draw();
					}
				}));
			}
		}

		
		/*const stack2 = new CircleStack({
			board, 
			cards: [
				new Card('hearts', '3'), 
				new Card('spades', 'ace'),
				new Card('clubs', 'king'),
				new Card('diamonds', '8'),
			],
			start: 2
		});*/
	}

	makeDummyHand();

	// board.draw() won't show anything until images load
	onLoaded(() => board.draw());

	rebuildPlayers();
	
	function rebuildPlayers() {
		for (let i=players.length; i<4; i++) {
			players[i] = 'Computer';
		}
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
				players.splice(i, 0, 'Computer');
				rebuildPlayers();
			});
			a.innerText = 'x';
			odiv.appendChild(pdiv);
			if (players[i] == 'Human' && i != 0) {
				odiv.appendChild(a);
			}
			div.appendChild(odiv);
		}
	}
	
	// Update board
	// Go clockwise from bottom
	function updateBoard(data) {
		playerId = data.Player;
		myActions = data.Actions;
		const nh = data.Hands.length;
		let lrtbs = ['bottom', 'left', 'top', 'right'];
		let offsets = [0, 0, 0, 0];
		// Rebuild hands if necessary
		if (board.hands.length != nh) {
			board.hands = [];

			for (let i=0; i<nh; i++) {
				const j = (playerId+i)%nh;
				const hand = new Hand({board, lrtb: lrtbs[i], name: data.Names[j], offset: offsets[i]});
			}
		}

		for (let i=0; i<nh; i++) {
			const j = (playerId+i)%nh;
			const hand = board.hands[i];

			// Update names as players join
			hand.name = data.Names[j];
			
			// Ditch holding
			if (hand.holding) {
				hand.holding = null;
			}
			
			// Update cards
			hand.cards = [];
			for (let k=0; k<data.Hands[j].length; k++) {
				const cardIdx = data.Hands[j][k];
				let suit, rank, card;
				if (cardIdx == -1) {
					suit = suits[0];
					rank = ranks[0];
					card = new Card(suit, rank);
					card.visible = false;
				} else {
					suit = suits[Math.floor(cardIdx/13)];
					rank = ranks[cardIdx % 13];
					card = new Card(suit, rank);
					card.visible = true;
				}
				// For switching visibility later
				card.cardIdx = cardIdx;
				// Don't show opponent cards
				if (i != 0) {
					card.visible = false;
				}
				hand.cards.push(card);
			}
		}

		// Update per-player icons and messages
		for (let i=0; i<nh; i++) {
			const j = (playerId+i)%nh;
			const hand = board.hands[i];

			hand.buttons = [];

			if (data.Bids[j] != -1) {
				hand.buttons.push(new Button({text: 'Bid: ' + data.Bids[j]}));
			}
			if (data.Bids[3] != -1) {
				hand.buttons.push(new Button({text: 'Tricks: ' + data.Tricks[j]}));
			}
		}
		
		// Update buttons
		const hand = board.hands[0];
		let bidv = 0;
		let bidb = null;
		for (let i=0; i<myActions.length; i++) {
			const act = myActions[i];
			if (act.Verb == verbs.indexOf("Bid")) {
				bidb = new Button({text: 'Bid: 0'});
				hand.buttons = [];
				hand.buttons.push(bidb);
				hand.buttons.push(new Button({
					text: '-',
					cb: () => {
						if (bidv > 0) {
							bidv--;
							bidb.text = 'Bid: ' + bidv;
							board.draw();
						}
					},
				}));
				hand.buttons.push(new Button({
					text: '+',
					cb: () => {
						if (bidv < 13) {
							bidv++;
							bidb.text = 'Bid: ' + bidv;
							board.draw();
						}
					},
				}));
				hand.buttons.push(new Button({
					text: 'Place Bid',
					cb: () => {
						act.Bid = bidv;
						conn.send(JSON.stringify({'Type': 'Action', 
							'Game': gameId, 
							'Data': JSON.stringify(act)}));
					}
				}));
				break;
			}
		}

		// Update stacks
		board.stacks = [];
		const cards = [];

		for (let i=0; i<4; i++) {
			const cardIdx = data.Trick[i];
			if (cardIdx >= 0) {
				const suit = suits[Math.floor(cardIdx/13)];
				const rank = ranks[cardIdx % 13];
				const card = new Card(suit, rank);
				card.visible = true;
				cards.push(card);
			}
		}

		const stack = new CircleStack({
			board, 
			cards: cards,
			start: (playerId + data.Attacker)%4,
		});

		board.stacks.push(stack);

		board.message = "";

		let tricks = 0;
		for (let i=0; i<4; i++) {
			tricks += data.Tricks[i];
		}

		// Show over message
		if (tricks == 13) {
			board.message = "Game over!";
		}

		// TODO: display old tricks
		
		board.draw();
	}
	
	conn = new WebSocket(`ws://${location.host}/ws`);

	/*conn.onopen = () => {
		conn.send(JSON.stringify({'Type': 'List'}));
	}*/

	conn.onmessage = e => {
		const json = JSON.parse(e.data);
		const data = json.Data ? JSON.parse(json.Data) : null;
		switch (json.Type) {
			case 'Error':
				alert(json.Data);
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

	function addPlayer() {
		for (let i=0; i<players.length; i++) {
			if (players[i] == 'Computer') {
				players[i] = 'Human';
				break;
			}
		}
		rebuildPlayers();
	}
	
	$('#human').addEventListener('click', e => {
		addPlayer();
	});

	canvas.addEventListener('mousemove', e => {
		board.mousemove(e);
	});

	canvas.addEventListener('mousedown', e => {
		board.mousedown(e);
	});
	
	canvas.addEventListener('mouseup', e => {
		const hand = board.hands[0];
		const point = new DOMPoint(e.offsetX, e.offsetY);
		if (hand && hand.holding) {
			// Try to play
			// Note that only bottom player will have option of playing a card so we don't need to theck x for left and right lrtb
			if (point.y > 100 && point.y <= canvas.height-100) {
				for (let i=0; i<myActions.length; i++) {
					const action = myActions[i];
					const card = cardToIndex(hand.holding);
					if (action.Card == card) {
						conn.send(JSON.stringify({'Type': 'Action', 'Game': gameId, 'Data': JSON.stringify(action)}));
					}
				}
			}
		}
		board.mouseup();
	});
	
	$('#start').addEventListener('click', () => {
		//makeDummyHand();
		conn.send(JSON.stringify({'Type': 'New', 'Types': players, 'Name': $('#name').value}));
	});

	$('#join').addEventListener('click', () => {
		//makeDummyHand();
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

});
