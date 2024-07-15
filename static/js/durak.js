import {loadCardImages, onLoaded, loadImage, Button, Card, Hand, Board, Stack, Deck} from './ui.js';
import {$, $$} from './util.js';

window.addEventListener('load', () => {
	loadCardImages();

	const suits = ["clubs", "spades", "hearts", "diamonds"];
	const ranks = ["6", "7", "8", "9", "10", "jack", "queen", "king", "ace"];
	const verbs = ["Play", "Cover", "Reverse", "Pass", "PickUp", "Defer"];

	function cardToIndex(card) {
		let i = suits.indexOf(card.suit);
		let j = ranks.indexOf(card.rank);
		if (i != -1 && j != -1) {
			return i*9+j;
		}
		return null;
	}

	const names = ['Emote', 'Artichoke', 'Json', 'Wrench', 'Ziggler', 'Fanny'];
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
		playerId = data.Player;
		myActions = data.Actions;
		const nh = data.Hands.length;
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
					suit = suits[Math.floor(cardIdx/9)];
					rank = ranks[cardIdx % 9];
					card = new Card(suit, rank);
					card.visible = true;
				}
				// For switching visibility later
				card.cardIdx = cardIdx;
				// Don't show even known enemy cards
				if (i != 0 && !$('#show-known').checked) {
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

			if (data.Attacker == j) {
				hand.buttons.push(new Button({img: swordImg}));
			} else if (data.Defender == j) {
				hand.buttons.push(new Button({img: shieldImg}));
			}

			if (data.PickingUp && data.Defender == j) {
				hand.buttons.push(new Button({text: 'Picking Up'}));
			} else if (data.Passed[j]) {
				hand.buttons.push(new Button({text: 'Passed'}));
			}
		}
		
		// Update buttons
		const hand = board.hands[0];
		for (let i=0; i<myActions.length; i++) {
			const act = myActions[i];
			if (act.Verb == verbs.indexOf("Pass")) {
				hand.buttons.push(new Button({
					text: 'Pass',
					cb: () => {
						conn.send(JSON.stringify({'Type': 'Action', 'Game': gameId, 'Data': JSON.stringify(act)}));
					},
				}));
			}
			if (myActions[i].Verb == verbs.indexOf("PickUp")) {
				hand.buttons.push(new Button({
					text: 'Pick Up',
					cb: () => {
						conn.send(JSON.stringify({'Type': 'Action', 'Game': gameId, 'Data': JSON.stringify(act)}));
					},
				}));
			} 
		}

		// Update stacks
		board.stacks = [];

		for (let i=0; i<data.Plays.length; i++) {
			const cardIdx = data.Plays[i];
			const suit = suits[Math.floor(cardIdx/9)];
			const rank = ranks[cardIdx % 9];
			const card = new Card(suit, rank);
			new Stack({board, cards: [card]});

			if (data.Covers[i] != -2) {
				const cardIdx = data.Covers[i];
				const suit = suits[Math.floor(cardIdx/9)];
				const rank = ranks[cardIdx % 9];
				const card = new Card(suit, rank);
				board.stacks.at(-1).cards.push(card);
			}
		}

		board.message = "";

		// Show win message
		if (data.Won[data.Player]) {
			board.message = "You won!";
		} else {
			// Show loss message
			let winners = 0;
			for (let i=0; i<data.Won.length; i++) {
				if (data.Won[i]) {
					winners++;
				}
			}
			if (winners == data.Won.length-1) {
				board.message = "You lost...";
			}
		}

		// Display deck and trump
		const cardIdx = data.Trump;
		const suit = suits[Math.floor(cardIdx/9)];
		const rank = ranks[cardIdx % 9];
		const trump = new Card(suit, rank);
		let p = {x: 100, y: 60};
		if (nh == 3) {
			p = {x: 100, y: 300};
		} 
		board.deck = new Deck({trump, p, size: data.CardsInDeck});

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
		const hand = board.hands[0];
		const point = new DOMPoint(e.offsetX, e.offsetY);
		if (hand && hand.holding) {
			// Check stacks
			let underCard = null;
			board.stacks.forEach(s => {
				const xfms = s.getCardTransforms();
				for (let i=0; i<s.cards.length; i++) {
					const p = point.matrixTransform(xfms[i].inverse());
					if (p.x > 0 && p.x < s.cards[i].width && p.y > 0 && p.y < s.cards[i].height) {
						underCard = s.cards[i];
					}
				}
			});
			if (underCard) {
				for (let i=0; i<myActions.length; i++) {
					const action = myActions[i];
					const underIdx = cardToIndex(underCard);
					const overIdx = cardToIndex(hand.holding);
					if (action.Card == overIdx && action.Covering == underIdx) {
						conn.send(JSON.stringify({'Type': 'Action', 'Game': gameId, 'Data': JSON.stringify(action)}));
					}
				}
			// Try to play or reverse
			// Note that only bottom player will have option of playing a card so we don't need to theck x for left and right lrtb
			} else if (point.y > 100 && point.y <= canvas.height-100) {
				for (let i=0; i<myActions.length; i++) {
					const action = myActions[i];
					const card = cardToIndex(hand.holding);
					if (action.Card == card && action.Covering == -2) {
						conn.send(JSON.stringify({'Type': 'Action', 'Game': gameId, 'Data': JSON.stringify(action)}));
					}
				}
			}
		}
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

	$('#show-known').addEventListener('change', e => {
		board.hands.forEach((h, idx) => {
			if (idx > 0) {
				h.cards.forEach(c => {
					c.visible = c.cardIdx != -1 && $('#show-known').checked;
				});
			}
		});
		board.draw();
	});
});
