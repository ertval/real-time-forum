import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import vm from "node:vm";
import { beforeAll, expect, test } from "vitest";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const repoRoot = path.resolve(__dirname, "..", "..", "..");

class MockClassList {
	constructor() {
		this.items = new Set();
	}

	add(...tokens) {
		for (const token of tokens) this.items.add(token);
	}

	remove(...tokens) {
		for (const token of tokens) this.items.delete(token);
	}

	contains(token) {
		return this.items.has(token);
	}

	toggle(token, force) {
		if (typeof force === "boolean") {
			if (force) {
				this.items.add(token);
			} else {
				this.items.delete(token);
			}
			return force;
		}
		if (this.items.has(token)) {
			this.items.delete(token);
			return false;
		}
		this.items.add(token);
		return true;
	}
}

class MockStyle {
	constructor() {
		this.values = new Map();
	}

	setProperty(name, value) {
		this.values.set(name, value);
	}

	removeProperty(name) {
		this.values.delete(name);
	}

	getPropertyValue(name) {
		return this.values.get(name) ?? "";
	}
}

class MockEventTarget {
	constructor() {
		this.listeners = new Map();
	}

	addEventListener(type, handler, options = {}) {
		if (!this.listeners.has(type)) this.listeners.set(type, []);
		const once = !!(options && typeof options === "object" && options.once);
		this.listeners.get(type).push({ handler, once });
	}

	removeEventListener(type, handler) {
		const items = this.listeners.get(type);
		if (!items) return;
		this.listeners.set(
			type,
			items.filter((item) => item.handler !== handler),
		);
	}

	dispatchEvent(event) {
		if (!event || typeof event.type !== "string") {
			throw new Error("event.type is required");
		}
		if (typeof event.preventDefault !== "function") {
			event.preventDefault = () => {};
		}
		if (typeof event.stopPropagation !== "function") {
			event.stopPropagation = () => {};
		}

		const items = this.listeners.get(event.type) ?? [];
		for (const item of [...items]) {
			item.handler.call(this, event);
			if (item.once) {
				this.removeEventListener(event.type, item.handler);
			}
		}
		return true;
	}
}

class MockElement extends MockEventTarget {
	constructor(tagName = "div") {
		super();
		this.tagName = tagName.toUpperCase();
		this.children = [];
		this.parentNode = null;
		this.classList = new MockClassList();
		this.style = new MockStyle();
		this.dataset = {};
		this.hidden = false;
		this.textContent = "";
		this.attributes = new Map();
	}

	appendChild(child) {
		child.parentNode = this;
		this.children.push(child);
		return child;
	}

	remove() {
		if (!this.parentNode) return;
		this.parentNode.children = this.parentNode.children.filter(
			(c) => c !== this,
		);
		this.parentNode = null;
	}

	setAttribute(name, value) {
		this.attributes.set(name, String(value));
	}

	getAttribute(name) {
		return this.attributes.get(name) ?? null;
	}

	removeAttribute(name) {
		this.attributes.delete(name);
	}

	getBoundingClientRect() {
		return { width: 0, height: 0 };
	}
}

class MockImageElement extends MockElement {
	constructor() {
		super("img");
		this._src = "";
		this.currentSrc = "";
		this.alt = "";
		this.complete = false;
		this.naturalWidth = 0;
		this.naturalHeight = 0;
	}

	set src(value) {
		const v = String(value);
		this._src = v;
		this.currentSrc = v;
		this.complete = true;
		if (!this.naturalWidth) this.naturalWidth = 40;
		if (!this.naturalHeight) this.naturalHeight = 40;
	}

	get src() {
		return this._src;
	}
}

class MockInputElement extends MockElement {
	constructor() {
		super("input");
		this.files = [];
		this._value = "";
	}

	set value(v) {
		this._value = String(v);
		if (this._value === "") {
			this.files = [];
		}
	}

	get value() {
		return this._value;
	}

	click() {
		this.dispatchEvent({ type: "click" });
	}
}

class MockCanvasContext {
	drawImage() {}

	getImageData() {
		const alpha = globalThis.__mockCanvasAlpha ?? 255;
		return { data: new Uint8ClampedArray([0, 0, 0, alpha]) };
	}
}

class MockCanvasElement extends MockElement {
	constructor() {
		super("canvas");
		this.width = 0;
		this.height = 0;
	}

	getContext() {
		return new MockCanvasContext();
	}
}

class MockDocument {
	constructor() {
		this.body = new MockElement("body");
	}

	createElement(tagName) {
		const tag = String(tagName).toLowerCase();
		if (tag === "canvas") return new MockCanvasElement();
		if (tag === "img") return new MockImageElement();
		if (tag === "input") return new MockInputElement();
		return new MockElement(tag);
	}

	querySelector() {
		return null;
	}
}

function installTestEnvironment() {
	const doc = new MockDocument();
	const win = {
		location: { href: "https://example.test/" },
		addEventListener() {},
		removeEventListener() {},
		setTimeout,
		clearTimeout,
	};

	globalThis.window = win;
	globalThis.document = doc;
	globalThis.Element = MockElement;
	globalThis.HTMLElement = MockElement;
	globalThis.HTMLImageElement = MockImageElement;

	const NativeURL = globalThis.URL;
	NativeURL.createObjectURL = () => "blob:mock-preview";
	NativeURL.revokeObjectURL = () => {};
	globalThis.URL = NativeURL;

	globalThis.__mockCanvasAlpha = 255;
}

function loadImagePickerForTests() {
	const filePath = path.join(
		repoRoot,
		"web",
		"static",
		"js",
		"image-picker.js",
	);
	const source = fs.readFileSync(filePath, "utf8");

	const noImports = source.replace(/^\s*import[\s\S]*?;\s*$/gm, "");

	const transformed = [
		"var MAX_IMAGE_BYTES = 20 * 1024 * 1024;",
		noImports.replace(
			"export function setupImagePicker",
			"function setupImagePicker",
		),
		"globalThis.__setupImagePicker = setupImagePicker;",
	].join("\n");

	vm.runInThisContext(transformed, { filename: "image-picker.test.eval.js" });
}

function loadPostsHelpersForTests() {
	const filePath = path.join(repoRoot, "web", "static", "js", "posts.js");
	const source = fs.readFileSync(filePath, "utf8");

	const noImports = source.replace(/^\s*import[\s\S]*?;\s*$/gm, "");

	const transformed = [
		noImports
			.replace(/export\s+async\s+function /g, "async function ")
			.replace(/export function /g, "function ")
			.replace(
				"function openImageLightbox(",
				"function __originalOpenImageLightbox(",
			)
			.replace(/\bopenImageLightbox\(/g, "globalThis.__openImageLightboxSpy("),
		"globalThis.__bindExpandableImage = bindExpandableImage;",
		"globalThis.__syncImageTransparencyPresentation = syncImageTransparencyPresentation;",
	].join("\n");

	vm.runInThisContext(transformed, { filename: "posts.test.eval.js" });
}

async function flushMicrotasks() {
	await Promise.resolve();
	await Promise.resolve();
	await new Promise((resolve) => setTimeout(resolve, 0));
}

beforeAll(() => {
	installTestEnvironment();
	loadImagePickerForTests();
	loadPostsHelpersForTests();
});

test("Image preview behavior", async () => {
	const setupImagePicker = globalThis.__setupImagePicker;
	expect(typeof setupImagePicker).toBe("function");

	const input = new MockInputElement();
	const triggerButton = new MockElement("button");
	const clearButton = new MockElement("button");
	clearButton.hidden = true;
	const nameLabel = new MockElement("span");
	const previewContainer = new MockElement("div");
	previewContainer.hidden = true;
	const previewImage = new MockImageElement();

	let tooLargeCalls = 0;

	setupImagePicker({
		input,
		triggerButton,
		clearButton,
		nameLabel,
		previewContainer,
		previewImage,
		maxBytes: 10,
		onTooLarge: () => {
			tooLargeCalls += 1;
		},
	});

	input.files = [{ name: "ok.jpg", size: 9, type: "image/jpeg" }];
	input.dispatchEvent({ type: "change" });

	expect(previewContainer.hidden).toBe(false);
	expect(clearButton.hidden).toBe(false);
	expect(nameLabel.textContent).toBe("Selected: ok.jpg");
	expect(previewImage.src.startsWith("blob:mock-preview")).toBeTruthy();

	input.value = "had-file";
	input.files = [{ name: "too-big.jpg", size: 11, type: "image/jpeg" }];
	input.dispatchEvent({ type: "change" });

	expect(tooLargeCalls).toBe(1);
	expect(input.value).toBe("");
	expect(input.files.length).toBe(0);
});

test("Preview checkerboard for transparent PNG", async () => {
	const setupImagePicker = globalThis.__setupImagePicker;

	const input = new MockInputElement();
	const previewContainer = new MockElement("div");
	previewContainer.hidden = true;
	const previewImage = new MockImageElement();

	setupImagePicker({
		input,
		previewContainer,
		previewImage,
		maxBytes: 1024,
	});

	globalThis.__mockCanvasAlpha = 80;

	input.files = [{ name: "transparent.png", size: 200, type: "image/png" }];
	input.dispatchEvent({ type: "change" });

	await flushMicrotasks();

	expect(previewContainer.hidden).toBe(false);
	expect(
		previewContainer.classList.contains("image-preview--checkerboard"),
	).toBeTruthy();
	expect(previewImage.dataset.transparent).toBe("true");
});

test("Posted image expand by click", () => {
	const bindExpandableImage = globalThis.__bindExpandableImage;
	expect(typeof bindExpandableImage).toBe("function");

	const calls = [];
	globalThis.__openImageLightboxSpy = (...args) => calls.push(args);

	const image = new MockImageElement();
	image.src = "https://example.test/uploads/post.jpg";
	image.alt = "Post image";
	image.dataset.transparent = "true";
	image.getBoundingClientRect = () => ({ width: 320.2, height: 180.7 });

	bindExpandableImage(image, "post");

	image.dispatchEvent({ type: "click" });

	expect(calls.length).toBe(1);
	expect(calls[0][0]).toBe(image.src);
	expect(calls[0][1]).toBe("Post image");
	expect(calls[0][2]).toBe("post");
	expect(calls[0][3]).toBe(true);
	expect(calls[0][4]).toEqual({ minWidth: 320, minHeight: 181 });
});

test("Posted transparent PNG checkerboard", async () => {
	const syncImageTransparencyPresentation =
		globalThis.__syncImageTransparencyPresentation;
	expect(typeof syncImageTransparencyPresentation).toBe("function");

	const frame = new MockElement("div");
	const image = new MockImageElement();
	image.src = "https://example.test/uploads/transparent.png";
	image.currentSrc = image.src;
	image.naturalWidth = 80;
	image.naturalHeight = 80;

	globalThis.__mockCanvasAlpha = 70;

	syncImageTransparencyPresentation({
		imgEl: image,
		frameEl: frame,
		checkerboardClass: "post-image--checkerboard",
	});

	await flushMicrotasks();

	expect(image.dataset.transparent).toBe("true");
	expect(frame.classList.contains("post-image--checkerboard")).toBeTruthy();
});
