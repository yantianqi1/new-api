/**
 * Starlight docs 内容集合 schema 的自包含副本。
 *
 * 背景：本机 Node 为 v26（nightly），拒绝为 node_modules 下的 .ts 文件做类型剥离，
 * 导致 `@astrojs/starlight/schema` 无法加载、集合丢失 frontmatter 默认值（draft/head 等），
 * 生产构建会过滤掉全部页面、dev 渲染崩溃。
 *
 * 本文件把 Starlight 所需的 frontmatter schema 复制到项目内（项目内 .ts 可正常剥离），
 * 仅有的差异是 hero 的 icon 字段不校验内置图标枚举（原实现依赖 node_modules 里的 Icons 表）。
 * 内容随 @astrojs/starlight 升级后需同步。
 */

import { z } from 'astro/zod';
import type { SchemaContext } from 'astro:content';

export const HeadConfigSchema = () =>
	z
		.array(
			z.object({
				tag: z.enum(['title', 'base', 'link', 'style', 'meta', 'script', 'noscript', 'template']),
				attrs: z.record(z.union([z.string(), z.boolean(), z.undefined()])).optional(),
				content: z.string().optional(),
			})
		)
		.default([]);

export const PrevNextLinkConfigSchema = () =>
	z
		.union([
			z.boolean(),
			z.string(),
			z
				.object({
					link: z.string().optional(),
					label: z.string().optional(),
				})
				.strict(),
		])
		.optional();

const tocDefaults = { minHeadingLevel: 2, maxHeadingLevel: 3 };

export const TableOfContentsSchema = () =>
	z
		.union([
			z.object({
				minHeadingLevel: z.number().int().min(1).max(6).optional().default(2),
				maxHeadingLevel: z.number().int().min(1).max(6).optional().default(3),
			}),
			z.boolean().transform((enabled) => (enabled ? tocDefaults : false)),
		])
		.default(tocDefaults)
		.refine((toc) => (toc ? toc.minHeadingLevel <= toc.maxHeadingLevel : true), {
			message: 'minHeadingLevel must be less than or equal to maxHeadingLevel',
		});

const badgeBaseSchema = z.object({
	variant: z.enum(['note', 'danger', 'success', 'caution', 'tip', 'default']).default('default'),
	class: z.string().optional(),
});

const badgeSchema = badgeBaseSchema.extend({ text: z.string() });
const i18nBadgeSchema = badgeBaseSchema.extend({
	text: z.union([z.string(), z.record(z.string())]),
});

export const BadgeConfigSchema = () =>
	z
		.union([z.string(), badgeSchema])
		.transform((badge) =>
			typeof badge === 'string' ? { variant: 'default' as const, text: badge } : badge
		)
		.optional();

export const I18nBadgeConfigSchema = () => z.union([z.string(), i18nBadgeSchema]).optional();

export const SidebarLinkItemHTMLAttributesSchema = () =>
	z
		.record(z.union([z.string(), z.number(), z.boolean(), z.undefined(), z.null()]))
		.default({});

export const HeroSchema = (_context: SchemaContext) =>
	z.object({
		title: z.string().optional(),
		tagline: z.string().optional(),
		image: z
			.union([
				z.object({ alt: z.string().default(''), file: z.string() }),
				z.object({
					alt: z.string().default(''),
					dark: z.string(),
					light: z.string(),
				}),
				z
					.object({ html: z.string() })
					.transform(({ html }) => ({ html, alt: '' })),
			])
			.optional(),
		actions: z
			.object({
				text: z.string(),
				link: z.string(),
				variant: z.enum(['primary', 'secondary', 'minimal']).default('primary'),
				icon: z.string().optional(),
				attrs: z.record(z.union([z.string(), z.number(), z.boolean()])).optional(),
			})
			.array()
			.default([]),
	});

export function StarlightFrontmatterSchema(_context: SchemaContext) {
	return z.object({
		title: z.string(),
		description: z.string().optional(),
		editUrl: z.union([z.string().url(), z.boolean()]).optional().default(true),
		head: HeadConfigSchema(),
		tableOfContents: TableOfContentsSchema().optional(),
		template: z.enum(['doc', 'splash']).default('doc'),
		hero: HeroSchema(_context).optional(),
		lastUpdated: z.union([z.date(), z.boolean()]).optional(),
		prev: PrevNextLinkConfigSchema(),
		next: PrevNextLinkConfigSchema(),
		sidebar: z
			.object({
				order: z.number().optional(),
				label: z.string().optional(),
				hidden: z.boolean().default(false),
				badge: BadgeConfigSchema(),
				attrs: SidebarLinkItemHTMLAttributesSchema(),
			})
			.default({}),
		banner: z
			.object({
				content: z.string(),
			})
			.optional(),
		pagefind: z.boolean().default(true),
		draft: z.boolean().default(false),
	});
}

export function docsSchema() {
	return (context: SchemaContext) => StarlightFrontmatterSchema(context);
}
