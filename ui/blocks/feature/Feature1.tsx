import React from 'react';
import ClientCompoment from '../../lib/ClientCompoment';
interface Feature1Props {
    title: string;
    description: string;
    mediaContent?: React.ReactNode; // 新增：允许传入自定义媒体内容（如视频播放器）
    isReversed?: boolean; // 控制图片和文字的位置
    features?: {
        icon: React.ReactNode;
        title: string;
        description: string;
    }[];
}

export const Feature1: React.FC<Feature1Props> = ({
    title,
    description,
    mediaContent,
    isReversed = false,
    features = [],
}) => {
    return (
        <div className="py-16">
            <div className="container mx-auto px-4">
                <div className={`flex flex-col ${isReversed ? 'md:flex-row-reverse' : 'md:flex-row'} items-center gap-12`}>
                    {/* 左侧媒体区域 */}
                    <div className="w-full md:w-1/2">
                        <div className="relative rounded-lg overflow-hidden shadow-lg">
                            <div className="w-full h-full">
                                {mediaContent}
                            </div>
                        </div>
                    </div>

                    {/* 右侧文字区域 */}
                    <div className="w-full md:w-1/2">
                        <h3 className="text-3xl font-bold mb-4">{title}</h3>
                        <p className="text-gray-600 mb-8">{description}</p>

                        {/* 特性列表 */}
                        {features.length > 0 && (
                            <div className="space-y-6">
                                {features.map((feature, index) => (
                                    <div key={index} className="flex items-start gap-4">
                                        <div className="flex-shrink-0 text-primary">
                                            {feature.icon}
                                        </div>
                                        <div>
                                            <h4 className="font-semibold text-lg">{feature.title}</h4>
                                            <p className="text-gray-500">{feature.description}</p>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};