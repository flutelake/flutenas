export default function chart_options_func(dark: boolean = false) {
	return {
		chart: {
			height: 260,
			type: 'line',
			toolbar: {
				show: false
			},
			zoom: {
				enabled: false
			}
		},
		stroke: {
			curve: 'smooth',
			width: 2
		},
		dataLabels: {
			enabled: false
		},
		series: [],
		xaxis: {
			type: 'datetime',
			categories: []
		},
		yaxis: {
			min: 0,
			// max: 100,
			labels: {
				formatter: (value: number) => `${value.toFixed(0)}%`
			}
		},
		grid: {
			strokeDashArray: 4
		},
		legend: {
			position: 'top',
			horizontalAlign: 'right'
		},
		tooltip: {
			x: {
				format: 'HH:mm:ss'
			}
		},
		theme: {
			mode: dark ? 'dark' : 'light'
		}
	};
}
