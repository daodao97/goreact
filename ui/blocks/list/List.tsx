import React from 'react';
import Masonry from "react-responsive-masonry";
import ResponsiveMasonry from "react-responsive-masonry";
import { Image } from '../../components/img/Image';

interface ListProps {
    title: string
    images: string[]
}

export const List: React.FC<ListProps> = ({ title, images }) => {
    return (
        <section className="max-w-[90rem] mx-auto w-full px-4 sm:px-8 lg:px-12 py-16 sm:py-20">
            {title && (
                <h2 className="text-4xl sm:text-6xl font-bold mb-8 tracking-normal">
                    {title}
                </h2>
            )}
            {/* @ts-ignore */}
            <Masonry columnsCount={4} gutter="10px">
                {images.map((image, i) => (
                    <Image
                        key={i}
                        src={image}
                        alt={`Image ${i}`}
                        className="w-full h-auto rounded-lg ring-1 ring-gray-200/50 hover:ring-2 hover:ring-gray-300/50 transition-all duration-300"
                    />
                ))}
            </Masonry>
        </section>

    )
}