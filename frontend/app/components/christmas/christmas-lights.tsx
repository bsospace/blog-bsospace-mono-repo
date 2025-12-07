"use client";

import React from "react";

export const ChristmasLights: React.FC<{ className?: string }> = ({ className }) => {
    return (
        <>
            <style jsx>{`
        @keyframes twinkle {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.3; }
        }

        .light {
          animation: twinkle 1.5s ease-in-out infinite;
        }

        .light:nth-child(2n) {
          animation-delay: 0.3s;
        }

        .light:nth-child(3n) {
          animation-delay: 0.6s;
        }

        .light:nth-child(4n) {
          animation-delay: 0.9s;
        }
      `}</style>
            <div className={`flex gap-4 ${className}`}>
                <div className="light w-3 h-3 rounded-full bg-red-500 shadow-lg shadow-red-500/50"></div>
                <div className="light w-3 h-3 rounded-full bg-green-500 shadow-lg shadow-green-500/50"></div>
                <div className="light w-3 h-3 rounded-full bg-yellow-500 shadow-lg shadow-yellow-500/50"></div>
                <div className="light w-3 h-3 rounded-full bg-blue-500 shadow-lg shadow-blue-500/50"></div>
                <div className="light w-3 h-3 rounded-full bg-purple-500 shadow-lg shadow-purple-500/50"></div>
                <div className="light w-3 h-3 rounded-full bg-pink-500 shadow-lg shadow-pink-500/50"></div>
            </div>
        </>
    );
};
