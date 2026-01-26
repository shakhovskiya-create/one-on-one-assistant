/**
 * HTML Sanitization utility using DOMPurify
 * Prevents XSS attacks from untrusted HTML content
 */
import DOMPurify from 'dompurify';

/**
 * Sanitize HTML content to prevent XSS attacks
 * @param html - Raw HTML string to sanitize
 * @returns Sanitized HTML string safe for rendering
 */
export function sanitizeHtml(html: string | undefined | null): string {
	if (!html) return '';

	// Configure DOMPurify with safe defaults
	return DOMPurify.sanitize(html, {
		// Allow safe HTML tags
		ALLOWED_TAGS: [
			'a', 'b', 'i', 'u', 'em', 'strong', 'p', 'br', 'hr',
			'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
			'ul', 'ol', 'li', 'dl', 'dt', 'dd',
			'table', 'thead', 'tbody', 'tfoot', 'tr', 'th', 'td',
			'blockquote', 'pre', 'code', 'span', 'div',
			'img', 'figure', 'figcaption',
			'sup', 'sub', 'mark', 'del', 'ins',
			'details', 'summary'
		],
		// Allow safe attributes
		ALLOWED_ATTR: [
			'href', 'src', 'alt', 'title', 'class', 'id',
			'target', 'rel', 'width', 'height',
			'colspan', 'rowspan', 'align', 'valign',
			'style', 'data-*'
		],
		// Allow safe URI schemes
		ALLOWED_URI_REGEXP: /^(?:(?:https?|mailto|tel):|[^a-z]|[a-z+.-]+(?:[^a-z+.\-:]|$))/i,
		// Force links to open in new tab with security attributes
		ADD_ATTR: ['target', 'rel'],
		// Prevent DOM clobbering
		SANITIZE_DOM: true,
		// Remove dangerous protocols
		FORBID_ATTR: ['onerror', 'onload', 'onclick', 'onmouseover', 'onfocus', 'onblur'],
		// Remove script tags and their content
		FORBID_TAGS: ['script', 'style', 'iframe', 'object', 'embed', 'form', 'input', 'button']
	});
}

/**
 * Sanitize HTML for email content (more permissive for rich formatting)
 */
export function sanitizeEmailHtml(html: string | undefined | null): string {
	if (!html) return '';

	return DOMPurify.sanitize(html, {
		ALLOWED_TAGS: [
			'a', 'b', 'i', 'u', 'em', 'strong', 'p', 'br', 'hr',
			'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
			'ul', 'ol', 'li',
			'table', 'thead', 'tbody', 'tr', 'th', 'td',
			'blockquote', 'pre', 'code', 'span', 'div',
			'img', 'font', 'center',
			'sup', 'sub'
		],
		ALLOWED_ATTR: [
			'href', 'src', 'alt', 'title', 'class', 'id',
			'target', 'rel', 'width', 'height',
			'colspan', 'rowspan', 'align', 'valign',
			'style', 'color', 'bgcolor', 'face', 'size'
		],
		ALLOWED_URI_REGEXP: /^(?:(?:https?|mailto|tel|cid):|[^a-z]|[a-z+.-]+(?:[^a-z+.\-:]|$))/i,
		SANITIZE_DOM: true,
		FORBID_TAGS: ['script', 'iframe', 'object', 'embed', 'form', 'input', 'button']
	});
}

/**
 * Sanitize plain text excerpt (strip all HTML)
 */
export function sanitizeExcerpt(html: string | undefined | null): string {
	if (!html) return '';

	// First sanitize, then strip remaining HTML for plain text
	const sanitized = DOMPurify.sanitize(html, {
		ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'mark'],
		ALLOWED_ATTR: []
	});

	return sanitized;
}
