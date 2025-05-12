import React, { useEffect, useState } from "react";
import { twMerge } from 'tailwind-merge'

type Button = {
    text: string;
    href: string;
    target?: string;
    rel?: string;
    className?: string;
    overwirteClassName?: string;
}

type Hero = {
    title: string;
    description: string;
    buttons: Button[];
    bgImage?: string;
    textPosition?: 'top' | 'center' | 'bottom';
    heightLevel?: 'full' | 'medium' | 'small';
}

export default function Hero({ hero, children }: { hero: Hero, children: React.ReactNode }) {
    const [windowHeight, setWindowHeight] = useState<string>("100vh");

    useEffect(() => {
        const updateHeight = () => setWindowHeight(`${window.innerHeight}px`);
        updateHeight();
        window.addEventListener('resize', updateHeight);
        return () => window.removeEventListener('resize', updateHeight);
    }, []);

    const getTextPositionClasses = () => {
        switch (hero.textPosition) {
            case 'top':
                return 'items-start pt-24';
            case 'bottom':
                return 'items-end pb-24';
            case 'center':
            default:
                return 'items-center';
        }
    };

    const getHeightStyle = () => {
        switch (hero.heightLevel) {
            case 'medium':
                return '70vh';
            case 'full':
                return windowHeight;
            case 'small':
            default:
                return '50vh';
        }
    };

    return (
        <section
            className={`relative px-4 sm:px-6 lg:px-8 shadow-xl flex ${getTextPositionClasses()} justify-center overflow-hidden`}
            style={{
                height: getHeightStyle(),
                backgroundImage: hero.bgImage ? `url(${hero.bgImage})` : 'none',
                backgroundSize: 'cover',
                backgroundPosition: 'center'
            }}
        >
            {!hero.bgImage && (
                <div className="absolute inset-0 bg-gradient-to-r from-indigo-600 to-purple-600 z-0"></div>
            )}

            {hero.bgImage && (
                <div className="absolute inset-x-0 bottom-0 h-1/5 bg-gradient-to-t from-gray-900/90 to-transparent z-0"></div>
            )}

            <div className="max-w-4xl mx-auto text-center relative z-10">
                <h1 className="text-4xl sm:text-5xl font-extrabold mb-6 tracking-tight text-white">
                    {hero.title}
                </h1>
                {hero.buttons && (
                    <div className="flex flex-wrap justify-center gap-3 mb-8">
                        {(hero.buttons || []).map((button, index) => (
                            <a
                                key={index}
                                href={button.href}
                                target={button.target}
                                rel={button.rel}
                                className={
                                    button.overwirteClassName ? button.overwirteClassName :
                                        twMerge(
                                            "bg-white/10 hover:bg-white/20 text-white px-4 py-2 rounded-full transition-all duration-300 hover:-translate-y-1",
                                        )
                                }
                            >
                                {button.text}
                            </a>
                        ))}
                    </div>
                )}
                <p className="text-xl text-white/90 mb-6 max-w-3xl mx-auto">
                    {hero.description}
                </p>

                {children}
            </div>
        </section>
    );
}
