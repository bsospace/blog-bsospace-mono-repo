"use client";

import React from "react";

interface SnowfallProps {
    count?: number;
    color?: string;
}

export const Snowfall: React.FC<SnowfallProps> = ({
    count = 50,
    color = "255,255,255"
}) => {
    const snowflakes = React.useMemo(() => {
        return Array.from({ length: count }).map((_, idx) => ({
            id: idx,
            start: (Math.random() * 100).toFixed(2),
            drift: (Math.random() * 30 - 15).toFixed(2),
            duration: (12 + Math.random() * 8).toFixed(2),
            delay: (Math.random() * 10).toFixed(2),
            size: (4 + Math.random() * 5).toFixed(2),
            opacity: (0.4 + Math.random() * 0.5).toFixed(2),
        }));
    }, [count]);

    return (
        <>
            <style jsx>{`
        @keyframes snowfall {
          0% {
            transform: translateX(var(--snow-x-start)) translateY(-10px);
            opacity: 0;
          }
          10% {
            opacity: var(--snow-opacity);
          }
          90% {
            opacity: var(--snow-opacity);
          }
          100% {
            transform: translateX(var(--snow-x-end)) translateY(100vh);
            opacity: 0;
          }
        }

        .snowflake {
          position: absolute;
          top: -10px;
          width: var(--snow-size);
          height: var(--snow-size);
          background: var(--snow-color);
          border-radius: 50%;
          animation: snowfall var(--snow-duration) linear var(--snow-delay) infinite;
          box-shadow: 0 0 10px rgba(255, 255, 255, 0.5);
        }
      `}</style>
            <div className="pointer-events-none fixed inset-0 overflow-hidden z-50" aria-hidden="true">
                {snowflakes.map((flake) => (
                    <span
                        key={flake.id}
                        className="snowflake"
                        style={{
                            ['--snow-x-start' as string]: `${flake.start}vw`,
                            ['--snow-x-end' as string]: `${parseFloat(flake.start) + parseFloat(flake.drift)}vw`,
                            ['--snow-duration' as string]: `${flake.duration}s`,
                            ['--snow-delay' as string]: `${flake.delay}s`,
                            ['--snow-size' as string]: `${flake.size}px`,
                            ['--snow-opacity' as string]: flake.opacity,
                            ['--snow-color' as string]: `rgba(${color},1)`,
                        } as React.CSSProperties}
                    />
                ))}
            </div>
        </>
    );
};
