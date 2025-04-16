import React, { useEffect, useRef, useState } from 'react';
import ReactPlayer from 'react-player';
import { ReactPlayerProps } from 'react-player';

interface VideoProps extends ReactPlayerProps {
    lazy?: boolean;
}

export const Video: React.FC<VideoProps> = ({ lazy = false, ...props }) => {
    const [shouldLoad, setShouldLoad] = useState(!lazy);
    const [isMobile, setIsMobile] = useState(false);
    const videoRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        // 检测是否为移动设备
        const checkMobile = () => {
            const userAgent = navigator.userAgent || navigator.vendor || (window as any).opera;
            const isMobileDevice = /android|iPad|iPhone|iPod|webOS|BlackBerry|IEMobile|Opera Mini/i.test(userAgent);
            setIsMobile(isMobileDevice);
        };

        checkMobile();
    }, []);

    useEffect(() => {
        if (!lazy || shouldLoad) return;

        const observer = new IntersectionObserver(
            ([entry]) => {
                if (entry.isIntersecting) {
                    setShouldLoad(true);
                    observer.disconnect();
                }
            },
            { threshold: 0.1 }
        );

        if (videoRef.current) {
            observer.observe(videoRef.current);
        }

        return () => observer.disconnect();
    }, [lazy, shouldLoad]);

    return (
        <div
            ref={videoRef}
            style={{
                width: props.width || '100%',
                height: props.height || '100%',
                background: '#000',
                position: 'relative'
            }}
        >
            {shouldLoad && (
                <ReactPlayer
                    width={props.width || '100%'}
                    height={props.height || '100%'}
                    playsinline={isMobile}
                    controls={true}
                    playing={true}
                    muted={true}
                    onError={(e) => {
                        console.error("视频播放错误:", e);
                    }}
                    config={{
                        file: {
                            attributes: {
                                playsInline: true,
                                preload: "auto",
                                controlsList: "nodownload",
                                disablePictureInPicture: isMobile,
                                crossOrigin: "anonymous"
                            },
                            forceVideo: true,
                        }
                    }}
                    {...props}
                />
            )}
        </div>
    );
};
