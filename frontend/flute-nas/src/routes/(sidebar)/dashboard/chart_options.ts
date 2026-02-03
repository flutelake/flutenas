// Mock chart options function
export default function chart_options_func(dark: boolean = false) {
  return {
    chart: {
      height: 350,
      type: 'heatmap',
    },
    dataLabels: {
      enabled: false
    },
    stroke: {
      width: 10
    },
    series: [],
    xaxis: {
      type: 'datetime',
      categories: [
        '2022-01-01',
        '2022-01-02',
        '2022-01-03',
        '2022-01-04',
        '2022-01-05',
        '2022-01-06',
        '2022-01-07',
        '2022-01-08',
        '2022-01-09'
      ],
    },
    tooltip: {
      x: {
        format: 'dd/MM/yy HH:mm'
      }
    },
    legend: {
      position: 'top',
      horizontalAlign: 'right'
    },
    theme: {
      mode: dark ? 'dark' : 'light'
    }
  };
}