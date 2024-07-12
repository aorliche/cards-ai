import {loadCardImages, onLoaded, loadImage, Button, Card, Hand, Board, Stack} from './ui.js';
import {$, $$} from './util.js';

window.addEventListener('load', () => {
	loadCardImages();

	const names = ['Emote', 'Artichoke', 'Json', 'Wrench', 'Ziggler', 'Fanny'];
	$('#name').value = names[Math.floor(names.length*Math.random())];

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
				console.log($('#games-select'));
				$('#games-select').appendChild(opt);
			}
		}
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
		conn.send(JSON.stringify({'Type': 'New', 'Types': players, 'Name': $('#name').value}));
	});

	$('#join').addEventListener('click', () => {
		const select = $('#games-select');
		const opt = select.options[select.selectedIndex];
		if (!opt) {
			return;
		}
		const key = parseInt(opt.innerText.slice(5));
		conn.send(JSON.stringify({'Type': 'Join', 'Game': key, 'Name': $('#name').value}));
	});

	$('#send').addEventListener('click', () => {
		conn.send(JSON.stringify({'Type': 'Chat', 'Game': gameId, 'Data': $('#message').value}));
	});

	const canvas = $('#board');
	const board = new Board({canvas});
	const swordImg = loadImage('/images/sword.png');
	const shieldImg = loadImage('/images/shield.png');
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
