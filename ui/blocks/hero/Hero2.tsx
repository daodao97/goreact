import React from 'react';

interface Hero2Props {
    title?: string;
    description?: string;
    ctaText?: string;
    ctaLink?: string;
    className?: string;
}

export default function Hero2({
    title,
    description,
    ctaText,
    ctaLink,
    className = ""
}: Hero2Props) {
    return (
        <div className={`text-center mb-16 ${className}`}>
            <div className="inline-block mb-6">
                <div className="relative">
                    <div className="absolute -inset-1 bg-gradient-to-r from-amber-500 via-yellow-400 to-amber-500 rounded-lg blur opacity-30"></div>
                    <div className="relative px-7 py-4 bg-white dark:bg-gray-800 ring-1 ring-amber-500/20 rounded-lg">
                        <svg className="w-12 h-12 text-amber-500 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                        </svg>
                    </div>
                </div>
            </div>

            <h1 className="text-5xl font-bold bg-clip-text text-transparent relative animate-shimmer bg-[length:200%_100%] bg-gradient-to-r from-yellow-600 via-amber-300 to-yellow-600 mb-6 drop-shadow">
                {title}
            </h1>

            <p className="text-xl text-gray-600 dark:text-gray-400 max-w-2xl mx-auto leading-relaxed">
                {description}
            </p>

            <div className="flex justify-center items-center gap-4 mt-12">
                <div className="h-px w-16 bg-gradient-to-r from-transparent via-amber-200 to-transparent"></div>
                <svg className="w-5 h-5 text-amber-400" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M12 15V3m0 12l-4-4m4 4l4-4M2 17l.621 2.485A2 2 0 004.561 21h14.878a2 2 0 001.94-1.515L22 17"></path>
                </svg>
                <div className="h-px w-16 bg-gradient-to-r from-transparent via-amber-200 to-transparent"></div>
            </div>

            <div className="mt-8 text-center">
                <a
                    href={ctaLink}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="group relative inline-flex flex-col items-center"
                >
                    <div className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-amber-600 hover:bg-amber-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-amber-500 transition-colors duration-200">
                        {ctaText}
                        <svg className="ml-2 -mr-1 w-4 h-4 group-hover:translate-x-1 transition-transform duration-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M14 5l7 7m0 0l-7 7m7-7H3"></path>
                        </svg>
                    </div>
                </a>
            </div>
        </div>
    );
}
