// place files you want to import through the `$lib` alias in this folder.

export function formatSpeed(bytesPerSecond: number): string {
    if (bytesPerSecond < 1024) {
        return `${bytesPerSecond.toFixed(2)} KB/s`;
    } else if (bytesPerSecond < 1024 * 1024) {
        return `${(bytesPerSecond / 1024).toFixed(2)} MB/s`;
    } else {
        return `${(bytesPerSecond / (1024 * 1024)).toFixed(2)} GB/s`;
    }
}

export function formatSize(bytesPerSecond: number): string {
    if (bytesPerSecond < 1024) {
        return `${bytesPerSecond.toFixed(2)} B`;
    } else if (bytesPerSecond < 1024 * 1024) {
        return `${(bytesPerSecond / 1024).toFixed(2)} KB`;
    } else if (bytesPerSecond < 1024 * 1024 * 1024) {
        return `${(bytesPerSecond / (1024 * 1024)).toFixed(2)} MB`;
    } else {
        return `${(bytesPerSecond / (1024 * 1024 * 1024)).toFixed(2)} GB`;
    }
}

export function formatDateTime(
    value: string | number | Date | null | undefined,
    locale: string = 'zh-CN'
): string {
    if (!value) {
        return '';
    }

    const date = value instanceof Date ? value : new Date(value);
    if (isNaN(date.getTime())) {
        return '';
    }

    const options: Intl.DateTimeFormatOptions = {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    };

    return date.toLocaleString(locale, options);
}
