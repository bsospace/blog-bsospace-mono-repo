"use client";

import React from "react";

export const ChristmasTree: React.FC<{ className?: string }> = ({ className }) => (
    <svg
        className={className}
        viewBox="0 0 200 260"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
    >
        <defs>
            <linearGradient id="treeGradLeft" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="#4ade80" />
                <stop offset="50%" stopColor="#22c55e" />
                <stop offset="100%" stopColor="#16a34a" />
            </linearGradient>
            <linearGradient id="treeGradRight" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="#16a34a" />
                <stop offset="50%" stopColor="#15803d" />
                <stop offset="100%" stopColor="#166534" />
            </linearGradient>
            <linearGradient id="trunkGrad" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="#a0522d" />
                <stop offset="100%" stopColor="#5d3a1a" />
            </linearGradient>
            <linearGradient id="starGrad" x1="0%" y1="0%" x2="0%" y2="100%">
                <stop offset="0%" stopColor="#fde047" />
                <stop offset="100%" stopColor="#eab308" />
            </linearGradient>
            <filter id="ornamentGlow" x="-100%" y="-100%" width="300%" height="300%">
                <feGaussianBlur stdDeviation="3" result="blur" />
                <feMerge>
                    <feMergeNode in="blur" />
                    <feMergeNode in="SourceGraphic" />
                </feMerge>
            </filter>
            <filter id="treeShadow" x="-50%" y="-50%" width="200%" height="200%">
                <feDropShadow dx="0" dy="4" stdDeviation="6" floodColor="#000" floodOpacity="0.3" />
            </filter>
            <radialGradient id="shadowGrad" cx="50%" cy="50%" r="50%">
                <stop offset="0%" stopColor="rgba(0,0,0,0.4)" />
                <stop offset="100%" stopColor="rgba(0,0,0,0)" />
            </radialGradient>
        </defs>

        <ellipse cx="100" cy="235" rx="60" ry="8" fill="url(#shadowGrad)" />
        <rect x="88" y="195" width="24" height="35" rx="2" fill="url(#trunkGrad)" />
        <path d="M100 25 L100 200 L35 200 Z" fill="url(#treeGradLeft)" filter="url(#treeShadow)" />
        <path d="M100 25 L100 200 L165 200 Z" fill="url(#treeGradRight)" filter="url(#treeShadow)" />

        <path
            d="M55 180 Q70 165, 100 170 Q130 175, 145 160"
            stroke="rgba(100,100,100,0.6)"
            strokeWidth="2"
            fill="none"
        />
        <path
            d="M50 140 Q75 125, 100 130 Q125 135, 140 120"
            stroke="rgba(100,100,100,0.6)"
            strokeWidth="2"
            fill="none"
        />
        <path
            d="M60 100 Q80 85, 100 90 Q120 95, 130 80"
            stroke="rgba(100,100,100,0.6)"
            strokeWidth="2"
            fill="none"
        />

        {/* Ornaments with animation */}
        <circle cx="60" cy="175" r="10" fill="#22d3ee" filter="url(#ornamentGlow)">
            <animate attributeName="opacity" values="0.9;1;0.9" dur="2s" repeatCount="indefinite" />
        </circle>
        <circle cx="90" cy="185" r="10" fill="#ef4444" filter="url(#ornamentGlow)">
            <animate attributeName="opacity" values="1;0.8;1" dur="2.5s" repeatCount="indefinite" />
        </circle>
        <circle cx="120" cy="180" r="10" fill="#fbbf24" filter="url(#ornamentGlow)">
            <animate attributeName="opacity" values="0.8;1;0.8" dur="1.8s" repeatCount="indefinite" />
        </circle>
        <circle cx="145" cy="165" r="10" fill="#f97316" filter="url(#ornamentGlow)">
            <animate attributeName="opacity" values="0.9;1;0.9" dur="2.2s" repeatCount="indefinite" />
        </circle>

        <circle cx="55" cy="135" r="9" fill="#fbbf24" filter="url(#ornamentGlow)">
            <animate attributeName="opacity" values="0.8;1;0.8" dur="2.3s" repeatCount="indefinite" />
        </circle>
        <circle cx="85" cy="125" r="9" fill="#22d3ee" filter="url(#ornamentGlow)">
            <animate attributeName="opacity" values="1;0.85;1" dur="1.9s" repeatCount="indefinite" />
        </circle>
        <circle cx="115" cy="130" r="9" fill="#ef4444" filter="url(#ornamentGlow)">
            <animate attributeName="opacity" values="0.85;1;0.85" dur="2.1s" repeatCount="indefinite" />
        </circle>
        <circle cx="138" cy="120" r="9" fill="#fbbf24" filter="url(#ornamentGlow)">
            <animate attributeName="opacity" values="0.9;1;0.9" dur="2.4s" repeatCount="indefinite" />
        </circle>

        <circle cx="65" cy="95" r="8" fill="#ef4444" filter="url(#ornamentGlow)">
            <animate attributeName="opacity" values="0.85;1;0.85" dur="2s" repeatCount="indefinite" />
        </circle>
        <circle cx="95" cy="85" r="8" fill="#fbbf24" filter="url(#ornamentGlow)">
            <animate attributeName="opacity" values="1;0.9;1" dur="1.7s" repeatCount="indefinite" />
        </circle>
        <circle cx="125" cy="90" r="8" fill="#22d3ee" filter="url(#ornamentGlow)">
            <animate attributeName="opacity" values="0.9;1;0.9" dur="2.2s" repeatCount="indefinite" />
        </circle>

        <circle cx="85" cy="60" r="7" fill="#22d3ee" filter="url(#ornamentGlow)">
            <animate attributeName="opacity" values="0.9;1;0.9" dur="1.8s" repeatCount="indefinite" />
        </circle>
        <circle cx="115" cy="65" r="7" fill="#fbbf24" filter="url(#ornamentGlow)">
            <animate attributeName="opacity" values="0.85;1;0.85" dur="2.1s" repeatCount="indefinite" />
        </circle>

        <polygon
            points="100,5 104,18 118,18 107,27 111,40 100,32 89,40 93,27 82,18 96,18"
            fill="url(#starGrad)"
            filter="url(#ornamentGlow)"
        >
            <animate attributeName="opacity" values="0.9;1;0.9" dur="1.5s" repeatCount="indefinite" />
        </polygon>

        {/* Gift boxes */}
        <rect x="20" y="215" width="35" height="30" rx="3" fill="#7c3aed" />
        <rect x="20" y="210" width="35" height="8" rx="2" fill="#8b5cf6" />
        <rect x="35" y="210" width="5" height="35" fill="#22d3ee" />
        <rect x="20" y="220" width="35" height="5" fill="#22d3ee" />
        <ellipse cx="32" cy="210" rx="6" ry="4" fill="#22d3ee" />
        <ellipse cx="43" cy="210" rx="6" ry="4" fill="#22d3ee" />
        <circle cx="37.5" cy="210" r="3" fill="#06b6d4" />

        <rect x="140" y="220" width="40" height="25" rx="3" fill="#f59e0b" />
        <rect x="140" y="215" width="40" height="8" rx="2" fill="#fbbf24" />
        <rect x="157" y="215" width="6" height="30" fill="#ef4444" />
        <rect x="140" y="225" width="40" height="5" fill="#ef4444" />
        <ellipse cx="153" cy="215" rx="6" ry="4" fill="#ef4444" />
        <ellipse cx="167" cy="215" rx="6" ry="4" fill="#ef4444" />
        <circle cx="160" cy="215" r="3" fill="#dc2626" />
    </svg>
);
