interface SIABProps {
  environment: string;
  mpi_command: string;
  abacus_command: string;
  //
  pseudo_dir: string;
  pseudo_name: string;
  ecutwfc: number;
  bessel_nao_smooth: number;
  bessel_nao_rcut: number[];
  smearing_sigma: number;

  //
  optimizer: 'pytorch.SWAT' | 'bfgs';
  max_steps: number;
  spill_coefs: number[];
  spill_guess: string;
  nthreads_rcut: number;
  jY_type: string;
  //
  reference_systems: {
    shape: string;
    nbands: number;
    nspin: number;
    bond_lengths: number[];
  }[];

  //
  orbitals: {
    zeta_notation: string;
    shape: string;
    nbands_ref: number;
    orb_ref: string;
  }[];
}

const SIAB: SIABProps = {
  /* 数值原子轨道生成代码简称为
    SIAB（Systematically Improvable Atomic Basis）
    这些内容在镜像中已被指定，因此不需要指定 */
  environment: '',
  mpi_command: 'mpirun -np 8',
  abacus_command: 'abacus',

  /* 这一部分支持 ABACUS INPUT 中的所有参数 */
  pseudo_dir:
    '/root/abacus-develop/pseudopotentials/sg15_oncv_upf_2020-02-06/1.0', // 这个路径在镜像中已被指定，因此不需要指定
  pseudo_name: 'Si_ONCV_PBE-1.0.upf', // 下拉菜单方式，选择赝势名称
  ecutwfc: 60, // 输入框方式，输入截断能
  bessel_nao_smooth: 0, // 框选方式，选择是否平滑
  bessel_nao_rcut: [6, 7, 8, 9, 10], // NOTE: 下拉菜单方式，选择截断半径(多选)
  smearing_sigma: 0.01,

  /* SIAB 计算的参数设置 */
  optimizer: 'bfgs', // 双选框方式，pytorch.SWAT, bfgs
  max_steps: 1000, // 输入框方式，输入最大步数
  spill_coefs: [0.0, 1.0], // 输入框方式，输入泄漏系数
  spill_guess: 'atomic', // 对于"optimizer": "pytorch.SWAT"，目前支持random和identity。对于"optimizer": "bfgs"，支持random和atomic。
  nthreads_rcut: 4,
  jY_type: 'reduced', // 仅对于"optimizer": "bfgs"有效

  reference_systems: [
    {
      shape: 'dimer', // dimer, trimer, tetrahedron, square, triangular_bipyramid, octahedron, cube
      nbands: 8, // auto
      nspin: 1,
      bond_lengths: [1.62, 1.82, 2.22, 2.72, 3.22],
    },
    {
      shape: 'trimer',
      nbands: 10,
      nspin: 1,
      bond_lengths: [1.9, 2.1, 2.6],
    },
  ],

  orbitals: [
    {
      zeta_notation: 'Z',
      shape: 'dimer',
      nbands_ref: 4,
      orb_ref: 'none',
    },
    {
      zeta_notation: 'DZP',
      shape: 'dimer',
      nbands_ref: 4,
      orb_ref: 'Z',
    },
    {
      zeta_notation: 'TZDP',
      shape: 'trimer',
      nbands_ref: 6,
      orb_ref: 'DZP',
    },
  ],
};
