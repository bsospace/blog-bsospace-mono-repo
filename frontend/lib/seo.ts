import { Metadata } from 'next';
import envConfig from '../app/configs/envConfig';

export interface SEOData {
  title: string;
  description: string;
  keywords?: string[];
  image?: string;
  url?: string;
  type?: 'website' | 'article' | 'profile';
  author?: string;
  publishedTime?: string;
  modifiedTime?: string;
  section?: string;
  tags?: string[];
  noindex?: boolean;
  nofollow?: boolean;
}

export function generateMetadata(data: SEOData): Metadata {
  const {
    title,
    description,
    keywords = [],
    image = `${envConfig.domain}/blog-image.webp`,
    url,
    type = 'website',
    author,
    publishedTime,
    modifiedTime,
    section,
    tags = [],
    noindex = false,
    nofollow = false,
  } = data;

  const fullTitle = title ? `${title} | ${envConfig.organizationName} Blog` : `${envConfig.organizationName} Blog - Software Engineering Knowledge Hub`;
  const fullDescription = description || 'BSO Blog is a collaborative blogging platform created by Software Engineering students, aimed at sharing knowledge, cutting-edge techniques, and real-world experiences.';

  const defaultKeywords = [
    'software engineering',
    'programming',
    'technology',
    'blog',
    'coding',
    'development',
    envConfig.organizationName,
    'student projects',
    'tech knowledge'
  ];

  const allKeywords = [...new Set([...defaultKeywords, ...keywords])];

  const metadata: Metadata = {
    title: fullTitle,
    description: fullDescription,
    keywords: allKeywords,
    authors: [{ name: author || `${envConfig.organizationName} Team` }],
    creator: envConfig.organizationName,
    publisher: envConfig.organizationName,
    robots: {
      index: !noindex,
      follow: !nofollow,
      googleBot: {
        index: !noindex,
        follow: !nofollow,
        'max-video-preview': -1,
        'max-image-preview': 'large',
        'max-snippet': -1,
      },
    },
    alternates: {
      canonical: url || envConfig.domain,
    },
    openGraph: {
      title: fullTitle,
      description: fullDescription,
      url: url || envConfig.domain,
      siteName: `${envConfig.organizationName} Blog`,
      images: [
        {
          url: image,
          width: 1200,
          height: 630,
          alt: fullTitle,
          type: 'image/webp',
        },
      ],
      locale: 'en_US',
      type: type,
    },
    twitter: {
      card: 'summary_large_image',
      title: fullTitle,
      description: fullDescription,
      images: [image],
      creator: '@bsospace',
      site: '@bsospace',
    },
  };

  // Add article-specific metadata
  if (type === 'article') {
    metadata.openGraph = {
      ...metadata.openGraph,
      type: 'article',
    };

    if (publishedTime) {
      metadata.openGraph = {
        ...metadata.openGraph,
        publishedTime,
      };
    }

    if (modifiedTime) {
      metadata.openGraph = {
        ...metadata.openGraph,
        modifiedTime,
      };
    }

    if (author) {
      metadata.openGraph = {
        ...metadata.openGraph,
        authors: [author],
      };
    }

    if (section) {
      metadata.openGraph = {
        ...metadata.openGraph,
        section,
      };
    }

    if (tags.length > 0) {
      metadata.openGraph = {
        ...metadata.openGraph,
        tags,
      };
    }
  }

  return metadata;
}

export function generateArticleStructuredData(data: SEOData) {
  const {
    title,
    description,
    image = `${envConfig.domain}/blog-image.webp`,
    url,
    type = 'website',
    author,
    publishedTime,
    modifiedTime,
  } = data;

  const fullTitle = title ? `${title} | ${envConfig.organizationName} Blog` : `${envConfig.organizationName} Blog - Software Engineering Knowledge Hub`;
  const fullDescription = description || 'BSO Blog is a collaborative blogging platform created by Software Engineering students, aimed at sharing knowledge, cutting-edge techniques, and real-world experiences.';
  const currentUrl = url || envConfig.domain;

  if (type === 'article') {
    return {
      "@context": "https://schema.org",
      "@type": "Article",
      "headline": fullTitle,
      "description": fullDescription,
      "image": image,
      "author": {
        "@type": "Person",
        "name": author || `${envConfig.organizationName} Team`
      },
      "publisher": {
        "@type": "Organization",
        "name": envConfig.organizationName,
        "logo": {
          "@type": "ImageObject",
          "url": `${envConfig.domain}/logo.png`
        }
      },
      "datePublished": publishedTime,
      "dateModified": modifiedTime || publishedTime,
      "mainEntityOfPage": {
        "@type": "WebPage",
        "@id": currentUrl
      }
    };
  }

  return {
    "@context": "https://schema.org",
    "@type": "WebPage",
    "name": fullTitle,
    "description": fullDescription,
    "url": currentUrl,
    "publisher": {
      "@type": "Organization",
      "name": envConfig.organizationName,
      "logo": {
        "@type": "ImageObject",
        "url": `${envConfig.domain}/logo.png`
      }
    }
  };
}

export function generateBreadcrumbStructuredData(breadcrumbs: Array<{ label: string; href?: string }>) {
  return {
    "@context": "https://schema.org",
    "@type": "BreadcrumbList",
    "itemListElement": [
      {
        "@type": "ListItem",
        "position": 1,
        "name": "Home",
        "item": `${envConfig.domain}/home`
      },
      ...breadcrumbs.map((item, index) => ({
        "@type": "ListItem",
        "position": index + 2,
        "name": item.label,
        "item": item.href ? `${envConfig.domain}${item.href}` : undefined
      }))
    ]
  };
}

export function generateOrganizationStructuredData() {
  return {
    "@context": "https://schema.org",
    "@type": "Organization",
    "name": envConfig.organizationName,
    "url": envConfig.domain,
    "logo": `${envConfig.domain}/logo.png`,
    "description": "Software Engineering Knowledge Hub - Collaborative blogging platform for students",
    "sameAs": [
      "https://twitter.com/bsospace",
      "https://github.com/bsospace"
    ],
    "contactPoint": {
      "@type": "ContactPoint",
      "contactType": "customer service",
      "email": envConfig.email
    }
  };
}

export function generateWebsiteStructuredData() {
  return {
    "@context": "https://schema.org",
    "@type": "WebSite",
    "name": `${envConfig.organizationName} Blog`,
    "description": "Software Engineering Knowledge Hub - Collaborative blogging platform for students",
    "url": envConfig.domain,
    "potentialAction": {
      "@type": "SearchAction",
      "target": `${envConfig.domain}/search?q={search_term_string}`,
      "query-input": "required name=search_term_string"
    }
  };
}

export function sanitizeText(text: string, maxLength: number = 160): string {
  // Remove HTML tags
  const cleanText = text.replace(/<[^>]*>/g, '');
  
  // Remove extra whitespace
  const trimmedText = cleanText.replace(/\s+/g, ' ').trim();
  
  // Truncate if too long
  if (trimmedText.length <= maxLength) {
    return trimmedText;
  }
  
  // Truncate at word boundary
  const truncated = trimmedText.substring(0, maxLength);
  const lastSpace = truncated.lastIndexOf(' ');
  
  return lastSpace > 0 ? truncated.substring(0, lastSpace) + '...' : truncated + '...';
}

export function generateSlug(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^\w\s-]/g, '') // Remove special characters
    .replace(/\s+/g, '-') // Replace spaces with hyphens
    .replace(/-+/g, '-') // Replace multiple hyphens with single hyphen
    .trim();
} 