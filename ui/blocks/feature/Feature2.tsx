import React, { useState, useEffect, useRef } from 'react';
import ClientCompoment from '../../lib/ClientCompoment';
interface Feature2Props {
    title: string;
    description: string;
    isReversed?: boolean; // 控制图片和文字的位置
    layout?: 'horizontal' | 'vertical'; // 新增布局参数：水平或垂直
    features?: {
        icon: React.ReactNode;
        title: string;
        description: string;
        mediaContent: React.ReactNode;
    }[];
    autoplay?: boolean; // 控制是否自动轮播
    interval?: number; // 轮播间隔时间(毫秒)
}

export const Feature2: React.FC<Feature2Props> = ({
    title,
    description,
    isReversed = false,
    layout = 'horizontal', // 默认水平布局
    features = [],
    autoplay = true, // 默认启用自动轮播
    interval = 5000, // 默认5秒切换一次
}) => {
    // 添加状态来跟踪当前选中的特性
    const [activeFeatureIndex, setActiveFeatureIndex] = useState(0);
    const [prevFeatureIndex, setPrevFeatureIndex] = useState(0);
    const [isTransitioning, setIsTransitioning] = useState(false);
    const autoplayTimerRef = useRef<number | null>(null);

    // 处理特性切换
    const handleFeatureChange = (index: number) => {
        if (index !== activeFeatureIndex && !isTransitioning) {
            setPrevFeatureIndex(activeFeatureIndex);
            setIsTransitioning(true);
            setActiveFeatureIndex(index);

            // 500ms后结束过渡状态
            setTimeout(() => {
                setIsTransitioning(false);
            }, 500);
        }
    };

    // 重置轮播定时器
    const resetAutoplayTimer = () => {
        if (autoplayTimerRef.current !== null) {
            window.clearInterval(autoplayTimerRef.current);
            autoplayTimerRef.current = null;
        }

        if (autoplay && features.length > 1) {
            autoplayTimerRef.current = window.setInterval(() => {
                setActiveFeatureIndex((prevIndex) => {
                    const nextIndex = (prevIndex + 1) % features.length;
                    setPrevFeatureIndex(prevIndex);
                    setIsTransitioning(true);

                    // 500ms后结束过渡状态
                    setTimeout(() => {
                        setIsTransitioning(false);
                    }, 500);

                    return nextIndex;
                });
            }, interval);
        }
    };

    // 处理自动轮播
    useEffect(() => {
        resetAutoplayTimer();

        // 清理函数
        return () => {
            if (autoplayTimerRef.current !== null) {
                window.clearInterval(autoplayTimerRef.current);
                autoplayTimerRef.current = null;
            }
        };
    }, [autoplay, features.length, interval]); // 移除 activeFeatureIndex 依赖

    // 用户交互时暂停自动轮播
    const handleFeatureClick = (index: number) => {
        // 如果正在过渡或点击当前活动项，则不执行操作
        if (isTransitioning || index === activeFeatureIndex) return;

        handleFeatureChange(index);
        resetAutoplayTimer();
    };

    // 确定滑动方向
    const getSlideDirection = () => {
        if (activeFeatureIndex === prevFeatureIndex) return '';

        // 处理循环情况
        if (prevFeatureIndex === features.length - 1 && activeFeatureIndex === 0) {
            return 'slide-right';
        }
        if (prevFeatureIndex === 0 && activeFeatureIndex === features.length - 1) {
            return 'slide-left';
        }

        return activeFeatureIndex > prevFeatureIndex ? 'slide-right' : 'slide-left';
    };

    return (
        <div className="py-16">
            <div className="container mx-auto px-4">
                {layout === 'horizontal' ? (
                    // 水平布局 - 保持原有样式
                    <div className={`flex flex-col ${isReversed ? 'md:flex-row-reverse' : 'md:flex-row'} items-center gap-12`}>
                        {/* 图片区域 */}
                        <div className="w-full md:w-1/2">
                            <div className="relative rounded-lg overflow-hidden shadow-lg h-[400px]">
                                {/* 图片容器 */}
                                <div className="absolute inset-0 w-full h-full flex items-center justify-center">
                                    {features.map((feature, index) => (
                                        <div
                                            key={index}
                                            className={`absolute inset-0 w-full h-full transition-all duration-500 ease-in-out flex items-center justify-center
                                                ${index === activeFeatureIndex ? 'opacity-100 z-10' : 'opacity-0 z-0'}
                                                ${index === activeFeatureIndex && isTransitioning ? getSlideDirection() : ''}
                                            `}
                                        >
                                            <div className="w-full h-full">
                                                <ClientCompoment fallback={
                                                    <div className="w-full h-full bg-gray-800 flex items-center justify-center">
                                                        <div className="animate-pulse">Loading...</div>
                                                    </div>
                                                }>
                                                    {feature.mediaContent}
                                                </ClientCompoment>
                                            </div>
                                        </div>
                                    ))}
                                </div>

                                {/* 轮播指示器 */}
                                {features.length > 1 && (
                                    <div className="absolute bottom-4 left-0 right-0 flex justify-center items-center z-20 w-full">
                                        <div className="flex items-center gap-2">
                                            {features.map((_, index) => (
                                                <button
                                                    key={index}
                                                    className={`w-2 h-2 rounded-full transition-all ${activeFeatureIndex === index ? 'bg-white w-4' : 'bg-white/50'}`}
                                                    onClick={() => handleFeatureClick(index)}
                                                    aria-label={`查看特性 ${index + 1}`}
                                                    disabled={isTransitioning}
                                                />
                                            ))}
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>

                        {/* 文字区域 */}
                        <div className="w-full md:w-1/2">
                            <h3 className="text-3xl font-bold mb-4">{title}</h3>
                            {description !== "" && (
                                <p className="text-gray-600 mb-8">{description}</p>
                            )}

                            {/* 特性列表 */}
                            {features.length > 0 && (
                                <div className="space-y-6">
                                    {features.map((feature, index) => (
                                        <div
                                            key={index}
                                            className={`flex items-start gap-4 p-3 rounded-lg cursor-pointer transition-colors duration-200 ${activeFeatureIndex === index
                                                ? 'bg-gray-800 text-white'
                                                : 'hover:bg-gray-700 hover:text-gray-100'
                                                }`}
                                            onClick={() => handleFeatureClick(index)}
                                        >
                                            <div className="flex-shrink-0 text-primary">
                                                {feature.icon}
                                            </div>
                                            <div>
                                                <h4 className="font-semibold text-lg">{feature.title}</h4>
                                                {feature.description !== "" && (
                                                    <p className={`${activeFeatureIndex === index ? 'text-gray-300' : 'text-gray-500'}`}>
                                                        {feature.description}
                                                    </p>
                                                )}
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>
                    </div>
                ) : (
                    // 垂直布局 - 三层结构
                    <div className="flex flex-col items-center gap-12">
                        {/* 顶部：标题和描述 */}
                        <div className="w-full text-center">
                            <h3 className="text-3xl font-bold mb-4">{title}</h3>
                            {description !== "" && (
                                <p className="text-gray-600 mb-8 max-w-3xl mx-auto">{description}</p>
                            )}
                        </div>

                        {/* 中间：媒体内容 */}
                        <div className="w-full">
                            <div className="relative rounded-lg overflow-hidden shadow-lg h-[400px]">
                                {/* 图片容器 */}
                                <div className="absolute inset-0 w-full h-full flex items-center justify-center">
                                    {features.map((feature, index) => (
                                        <div
                                            key={index}
                                            className={`absolute inset-0 w-full h-full transition-all duration-500 ease-in-out flex items-center justify-center
                                                ${index === activeFeatureIndex ? 'opacity-100 z-10' : 'opacity-0 z-0'}
                                                ${index === activeFeatureIndex && isTransitioning ? getSlideDirection() : ''}
                                            `}
                                        >
                                            <div className="w-full h-full">
                                                <ClientCompoment fallback={
                                                    <div className="w-full h-full bg-gray-800 flex items-center justify-center">
                                                        <div className="animate-pulse">Loading...</div>
                                                    </div>
                                                }>
                                                    {feature.mediaContent}
                                                </ClientCompoment>
                                            </div>
                                        </div>
                                    ))}
                                </div>

                                {/* 轮播指示器 */}
                                {features.length > 1 && (
                                    <div className="absolute bottom-4 left-0 right-0 flex justify-center items-center z-20 w-full">
                                        <div className="flex items-center gap-2">
                                            {features.map((_, index) => (
                                                <button
                                                    key={index}
                                                    className={`w-2 h-2 rounded-full transition-all ${activeFeatureIndex === index ? 'bg-white w-4' : 'bg-white/50'}`}
                                                    onClick={() => handleFeatureClick(index)}
                                                    aria-label={`查看特性 ${index + 1}`}
                                                    disabled={isTransitioning}
                                                />
                                            ))}
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>

                        {/* 底部：横向排列的特性列表 - 单行布局 */}
                        {features.length > 0 && (
                            <div className="w-full overflow-x-auto">
                                <div className="flex flex-col md:flex-row flex-wrap md:flex-nowrap gap-6 min-w-min pb-4 justify-center">
                                    {features.map((feature, index) => (
                                        <div
                                            key={index}
                                            className={`flex-shrink-0 flex flex-col items-center text-center p-4 rounded-lg cursor-pointer transition-colors duration-200 w-full md:w-64 ${activeFeatureIndex === index
                                                ? 'bg-gray-800 text-white'
                                                : 'hover:bg-gray-700 hover:text-gray-100'
                                                }`}
                                            onClick={() => handleFeatureClick(index)}
                                        >
                                            <div className="flex-shrink-0 text-primary mb-3 text-2xl">
                                                {feature.icon}
                                            </div>
                                            <h4 className="font-semibold text-lg mb-2">{feature.title}</h4>
                                            {feature.description !== "" && (
                                                <p className={`${activeFeatureIndex === index ? 'text-gray-300' : 'text-gray-500'}`}>
                                                    {feature.description}
                                                </p>
                                            )}
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
};