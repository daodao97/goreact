import React, { useEffect, useState, useRef } from "react";
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
    bgVideo?: string;
    bgVideos?: string[]; // 支持多个视频URL
    videoInterval?: number; // 视频切换间隔，默认为5000ms
}

export default function Hero({ hero, children }: { hero: Hero, children: React.ReactNode }) {
    const [windowHeight, setWindowHeight] = useState<string>("100vh");
    const [currentVideoIndex, setCurrentVideoIndex] = useState(0);
    const [isTransitioning, setIsTransitioning] = useState(false);
    const [videos, setVideos] = useState<string[]>([]);
    const [isVideoReady, setIsVideoReady] = useState(false);
    const intervalRef = useRef<number | null>(null);

    // 处理视频数组
    useEffect(() => {
        if (hero.bgVideos && hero.bgVideos.length > 0) {
            setVideos(hero.bgVideos);
        } else if (hero.bgVideo) {
            setVideos([hero.bgVideo]);
        } else {
            setVideos([]);
        }
    }, [hero.bgVideo, hero.bgVideos]);

    // 窗口高度更新
    useEffect(() => {
        const updateHeight = () => setWindowHeight(`${window.innerHeight}px`);
        updateHeight();
        window.addEventListener('resize', updateHeight);
        return () => window.removeEventListener('resize', updateHeight);
    }, []);

    // 延迟设置视频为可加载状态
    useEffect(() => {
        // 使用一个小的延迟，确保关键内容优先加载
        const timer = setTimeout(() => {
            setIsVideoReady(true);
        }, 300); // 可以根据需要调整延迟时间
        return () => clearTimeout(timer);
    }, []);

    // 视频轮播定时器
    useEffect(() => {
        if (videos.length <= 1) return;

        const startCarousel = () => {
            if (intervalRef.current) {
                clearInterval(intervalRef.current);
            }

            // @ts-ignore
            intervalRef.current = setInterval(() => {
                setIsTransitioning(true);
                setTimeout(() => {
                    setCurrentVideoIndex((prev) => (prev + 1) % videos.length);
                    setIsTransitioning(false);
                }, 500); // 淡出动画时间
            }, hero.videoInterval || 5000);
        };

        startCarousel();

        return () => {
            if (intervalRef.current) {
                clearInterval(intervalRef.current);
            }
        };
    }, [videos, hero.videoInterval]);

    return (
        <div className="w-full h-full relative" style={{ width: '100%' }}>
            <section
                className="absolute left-0 right-0 shadow-xl flex items-center justify-center overflow-hidden"
                style={{
                    height: windowHeight,
                    width: '100vw',
                    margin: '0',
                    padding: '0',
                    position: 'relative',
                    left: '50%',
                    transform: 'translateX(-50%)',
                }}
            >
                {videos.length > 0 ? (
                    <>
                        {videos.map((videoUrl, index) => (
                            <video
                                key={index}
                                autoPlay
                                loop
                                muted
                                playsInline
                                poster={hero.bgImage}
                                preload="metadata"
                                className={`absolute inset-0 object-cover z-0 transition-opacity duration-500 ease-in-out ${index === currentVideoIndex ? 'opacity-100' : 'opacity-0'
                                    } ${isTransitioning && index === currentVideoIndex ? 'opacity-0' : ''}`}
                                style={{
                                    width: '100vw',
                                    height: '100%',
                                    objectFit: 'cover',
                                    position: 'absolute',
                                    left: 0,
                                    top: 0,
                                }}
                            >
                                {isVideoReady && <source src={videoUrl} type="video/mp4" />}
                            </video>
                        ))}
                    </>
                ) : !hero.bgImage ? (
                    <div className="absolute inset-0 bg-gradient-to-r from-indigo-600 to-purple-600 z-0"
                        style={{
                            width: '100vw',
                            height: '100%',
                            position: 'absolute',
                            left: 0,
                            top: 0,
                        }}
                    ></div>
                ) : (
                    <div
                        className="absolute inset-0 z-0"
                        style={{
                            backgroundImage: `url(${hero.bgImage})`,
                            backgroundSize: 'cover',
                            backgroundPosition: 'center',
                            width: '100vw',
                            height: '100%',
                            position: 'absolute',
                            left: 0,
                            top: 0,
                        }}
                    />
                )}

                {(hero.bgImage || videos.length > 0) && (
                    <div className="absolute inset-x-0 bottom-0 h-1/5 bg-gradient-to-t from-gray-900/90 to-transparent z-0"
                        style={{
                            width: '100vw',
                            position: 'absolute',
                            left: 0,
                        }}
                    ></div>
                )}

                <div className="w-full text-center relative z-10 pt-16">
                    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
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
                </div>
            </section>
        </div>
    );
}
