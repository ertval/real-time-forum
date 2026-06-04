import { fetchRoster } from './chat.roster.api.js';
import { renderRosterList } from './chat.roster.views.js';

const ROSTER_BOUND_ATTR = 'data-roster-bound';
const ROSTER_ROOT_SELECTOR = '[data-chat-roster]';
const ROSTER_LIST_ID = 'roster-list';
const ROW_SELECTOR = '[data-roster-user-id]';

// D04 will replace this one-shot painter with a reactive store driven by
// presence.update WebSocket events. Until then, a single Proxy is overkill.
function selectRow(rosterRoot, row) {
	const previouslySelected = rosterRoot.querySelectorAll('[aria-pressed="true"]');
	for (const node of previouslySelected) {
		node.setAttribute('aria-pressed', 'false');
		node.classList?.remove('is-selected');
	}

	row.setAttribute('aria-pressed', 'true');
	row.classList?.add('is-selected');
}

function emitSelection(row) {
	const userId = Number(row.getAttribute('data-roster-user-id') ?? 0);
	const username = row.getAttribute('data-roster-username') ?? '';
	const isOnline = row.getAttribute('data-roster-online') === 'true';

	row.dispatchEvent(
		new CustomEvent('chat:user-selected', {
			detail: { userId, username, isOnline },
			bubbles: true,
		}),
	);
}

function handleActivation(rosterRoot, target) {
	const row = target?.closest?.(ROW_SELECTOR);
	if (!row) {
		return;
	}

	selectRow(rosterRoot, row);
	emitSelection(row);
}

export function initChatRoster(options = {}) {
	const windowRef = options.windowRef ?? (typeof window !== 'undefined' ? window : null);
	const documentRef = options.documentRef ?? (typeof document !== 'undefined' ? document : null);
	const fetchRef = options.fetchRef ?? (typeof fetch === 'function' ? fetch : null);

	if (
		!windowRef ||
		!documentRef ||
		typeof fetchRef !== 'function' ||
		typeof documentRef.querySelector !== 'function'
	) {
		return null;
	}

	const rosterRoot = documentRef.querySelector(ROSTER_ROOT_SELECTOR);
	if (!rosterRoot) {
		return null;
	}

	if (rosterRoot.getAttribute(ROSTER_BOUND_ATTR) === 'true') {
		return null;
	}

	rosterRoot.setAttribute(ROSTER_BOUND_ATTR, 'true');

	rosterRoot.addEventListener('click', (event) => {
		handleActivation(rosterRoot, event.target);
	});

	rosterRoot.addEventListener('keydown', (event) => {
		if (event.key !== 'Enter' && event.key !== ' ' && event.key !== 'Spacebar') {
			return;
		}
		const row = event.target?.closest?.(ROW_SELECTOR);
		if (!row) {
			return;
		}
		event.preventDefault?.();
		selectRow(rosterRoot, row);
		emitSelection(row);
	});

	void (async () => {
		const entries = await fetchRoster(fetchRef);
		const listMount = rosterRoot.querySelector?.(`#${ROSTER_LIST_ID}`);
		if (!listMount) {
			return;
		}
		listMount.innerHTML = renderRosterList(entries);
	})();

	return { rosterRoot };
}
