// 定义事件监听器的类型
type EventCallback = (data: any) => void;

// 定义监听器映射接口
interface EventListeners {
    [event: string]: EventCallback[];
}

class EventBus {
    private listeners: EventListeners;

    constructor() {
        this.listeners = {};
    }

    on(event: string, callback: EventCallback): void {
        if (!this.listeners[event]) {
            this.listeners[event] = [];
        }
        this.listeners[event].push(callback);
    }

    off(event: string, callback: EventCallback): void {
        if (this.listeners[event]) {
            this.listeners[event] = this.listeners[event].filter(
                (listener) => listener !== callback
            );
        }
    }

    emit(event: string, data: any): void {
        if (this.listeners[event]) {
            this.listeners[event].forEach((listener) => listener(data));
        }
    }

}

const eventBus = new EventBus();
export default eventBus;