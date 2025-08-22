"use client";
import React, { createContext, useContext, ReactNode } from 'react';
import { Helmet } from 'react-helmet-async';
import envConfig from '../configs/envConfig';

interface SEOProps {
    title?: string;
    description?: string;
    image?: string;
    url?: string;
    keywords?: string;
    author?: string;
    robots?: string;
    twitterCardType?: 'summary' | 'summary_large_image';
    canonical?: string;
}

const defaultSEO: SEOProps = {
    title: 'BSO Blog',
    description: 'BSO Blog is a collaborative blogging platform created by Software Engineering students, aimed at sharing knowledge, cutting-edge techniques, and real-world experiences in software development, programming, and technology.',
    image: '/favicon.ico',
    url: envConfig.domain,
    keywords: 'BSO, Blog, Software Engineering, Dev, บทความเทคโนโลยี',
    author: `${envConfig.organizationName} Team`,
    robots: 'index, follow',
    twitterCardType: 'summary_large_image',
    canonical: envConfig.domain,
};

const SEOContext = createContext<SEOProps>(defaultSEO);

export const SEOProvider = ({ children, value }: { children: ReactNode; value?: SEOProps }) => {
    const seo = { ...defaultSEO, ...value };

    return (
        <SEOContext.Provider value={seo}>
            <Helmet>
                <title>BSO Blog | {seo.title}</title>
                <meta name="description" content={seo.description} />
                <meta name="keywords" content={seo.keywords} />
                <meta name="author" content={seo.author} />
                <meta name="robots" content={seo.robots} />
                <link rel="canonical" href={seo.canonical || seo.url} />

                {/* Open Graph Tags */}
                <meta property="og:title" content={seo.title} />
                <meta property="og:description" content={seo.description} />
                <meta property="og:image" content={seo.image} />
                <meta property="og:url" content={seo.url} />
                <meta property="og:type" content="website" />

                {/* Twitter Cards */}
                <meta name="twitter:card" content={seo.twitterCardType || 'summary_large_image'} />
                <meta name="twitter:title" content={seo.title} />
                <meta name="twitter:description" content={seo.description} />
                <meta name="twitter:image" content={seo.image} />

                {/* Favicon (optional, you may already have this in public/index.html) */}
                <link rel="icon" href="/favicon.ico" />
            </Helmet>
            {children}
        </SEOContext.Provider>
    );
};

export const useSEO = () => useContext(SEOContext);
