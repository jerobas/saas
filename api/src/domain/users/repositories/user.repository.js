export class UserRepository {
  constructor(dataSource) {
    this.dataSource = dataSource;
    this.repository = dataSource.getRepository('User');
  }

  async findById(id) {
    return this.repository.findOne({ where: { id } });
  }

  async findByEmail(email) {
    return this.repository.findOne({ where: { email } });
  }

  async findByAbacatePayCustomerId(abacatePayCustomerId) {
    return this.repository.findOne({ where: { abacatePayCustomerId } });
  }

  async create(data) {
    const user = this.repository.create(data);
    return this.repository.save(user);
  }

  async update(id, data) {
    await this.repository.update(id, data);
    return this.findById(id);
  }

  async delete(id) {
    return this.repository.delete(id);
  }
}
