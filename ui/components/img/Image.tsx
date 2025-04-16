import React from 'react';
import { PhotoProvider, PhotoView } from 'react-photo-view';

interface ImageProps {
    src: string;
    alt: string;
    width?: number;
    height?: number;
    className?: string;
}

export const Image: React.FC<ImageProps> = ({ src, alt, width = 600, height = "auto", className = '' }) => {
    return (
        <PhotoProvider>
            <PhotoView src={src}>
                <div className={`w-full h-full flex items-center justify-center ${className}`}>
                    <img
                        src={src}
                        alt={alt}
                        width={width}
                        height={height}
                        className={className}
                        onError={(e) => {
                            e.currentTarget.onerror = null; // 防止无限循环
                            e.currentTarget.src = "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='600' height='400' viewBox='0 0 600 400'%3E%3Crect width='100%25' height='100%25' fill='%23f3f4f6'/%3E%3Ctext x='50%25' y='50%25' dominant-baseline='middle' text-anchor='middle' font-family='sans-serif' font-size='18' fill='%236b7280'%3E图片加载失败%3C/text%3E%3C/svg%3E";
                            const imgElement = e.currentTarget;
                            imgElement.classList.add("bg-gray-100", "border", "border-gray-200");
                        }}
                    />
                </div>
            </PhotoView>
        </PhotoProvider>
    );
}

